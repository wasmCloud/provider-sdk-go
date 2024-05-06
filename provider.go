package provider

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type WasmcloudProvider struct {
	Id string

	context context.Context
	cancel  context.CancelFunc
	Logger  logr.Logger

	hostData HostData
	Topics   Topics

	natsConnection    *nats.Conn
	natsSubscriptions map[string]*nats.Subscription

	healthMsgFunc func() string

	shutdownFunc func() error
	shutdown     chan struct{}

	putSourceLinkFunc func(InterfaceLinkDefinition) error
	putTargetLinkFunc func(InterfaceLinkDefinition) error
	delSourceLinkFunc func(InterfaceLinkDefinition) error
	delTargetLinkFunc func(InterfaceLinkDefinition) error

	lock sync.Mutex
	// Links from the provider to other components, aka where the provider is the
	// source of the link. Indexed by the component ID of the target
	sourceLinks map[string]InterfaceLinkDefinition
	// Links from other components to the provider, aka where the provider is the
	// target of the link. Indexed by the component ID of the source
	targetLinks map[string]InterfaceLinkDefinition
}

func New(options ...ProviderHandler) (*WasmcloudProvider, error) {
	reader := bufio.NewReader(os.Stdin)

	// Make a channel to receive the host data so we can timeout if we don't receive it
	// All host data is sent immediately after the provider starts
	hostDataChannel := make(chan string, 1)
	go func() {
		hostDataRaw, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		hostDataChannel <- hostDataRaw
	}()

	hostData := HostData{}
	select {
	case hostDataRaw := <-hostDataChannel:
		err := json.Unmarshal([]byte(hostDataRaw), &hostData)
		if err != nil {
			return nil, err
		}
	case <-time.After(5 * time.Second):
		panic("failed to read host data, did not receive after 5 seconds")
	}

	// Initialize Logging
	logrusLog := logrus.New()
	if hostData.StructuredLogging {
		logrusLog.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrusLog.SetFormatter(&logrus.TextFormatter{})
	}

	// Connect to NATS
	nc, err := nats.Connect(hostData.LatticeRPCURL)
	if err != nil {
		return nil, err
	}

	// partition links based on if the provider is the source or target
	sourceLinks := []InterfaceLinkDefinition{}
	targetLinks := []InterfaceLinkDefinition{}

	// Loop over the numbers
	for _, link := range hostData.LinkDefinitions {
		if link.SourceID == hostData.ProviderKey {
			sourceLinks = append(sourceLinks, link)
		} else if link.Target == hostData.ProviderKey {
			targetLinks = append(targetLinks, link)
		} else {
			logrusLog.Warnf("Link %s->%s is not connected to provider, ignoring", link.SourceID, link.Target)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	provider := &WasmcloudProvider{
		Id: hostData.ProviderKey,

		context: ctx,
		cancel:  cancel,
		Logger: logrusr.New(logrusLog).
			WithName(hostData.HostID),

		hostData: hostData,
		Topics:   LatticeTopics(hostData),

		natsConnection:    nc,
		natsSubscriptions: map[string]*nats.Subscription{},

		healthMsgFunc: func() string { return "healthy" },

		shutdownFunc: func() error { return nil },
		shutdown:     make(chan struct{}),

		putSourceLinkFunc: func(InterfaceLinkDefinition) error { return nil },
		putTargetLinkFunc: func(InterfaceLinkDefinition) error { return nil },
		delSourceLinkFunc: func(InterfaceLinkDefinition) error { return nil },
		delTargetLinkFunc: func(InterfaceLinkDefinition) error { return nil },

		sourceLinks: make(map[string]InterfaceLinkDefinition, len(sourceLinks)),
		targetLinks: make(map[string]InterfaceLinkDefinition, len(targetLinks)),
	}

	for _, opt := range options {
		err := opt(provider)
		if err != nil {
			return nil, err
		}
	}

	for _, link := range sourceLinks {
		provider.putLink(link)
	}
	for _, link := range targetLinks {
		provider.putLink(link)
	}

	return provider, nil
}

func (wp *WasmcloudProvider) HostData() HostData {
	return wp.hostData
}

func (wp *WasmcloudProvider) Start() error {
	err := wp.subToNats()
	if err != nil {
		return err
	}

	wp.Logger.Info("provider started", "id", wp.Id)
	<-wp.context.Done()
	return nil
}

func (wp *WasmcloudProvider) subToNats() error {
	// ------------------ Subscribe to Health topic --------------------
	health, err := wp.natsConnection.Subscribe(wp.Topics.LATTICE_HEALTH,
		func(m *nats.Msg) {
			hc := HealthCheckResponse{
				Healthy: true,
				Message: "healthy",
			}

			hcBytes, err := json.Marshal(hc)
			if err != nil {
				wp.Logger.Error(err, "failed to encode health check")
				return
			}

			wp.natsConnection.Publish(m.Reply, hcBytes)
		})
	if err != nil {
		wp.Logger.Error(err, "LATTICE_HEALTH")
		return err
	}
	wp.natsSubscriptions[wp.Topics.LATTICE_HEALTH] = health

	// ------------------ Subscribe to Delete link topic --------------
	linkDel, err := wp.natsConnection.Subscribe(wp.Topics.LATTICE_LINK_DEL,
		func(m *nats.Msg) {
			link := InterfaceLinkDefinition{}
			err := json.Unmarshal(m.Data, &link)
			if err != nil {
				wp.Logger.Error(err, "failed to decode link")
				return
			}

			err = wp.deleteLink(link)
			if err != nil {
				// TODO(#10): handle better?
				wp.Logger.Error(err, "failed to delete link")
				return
			}
		})
	if err != nil {
		wp.Logger.Error(err, "LINK_DEL")
		return err
	}
	wp.natsSubscriptions[wp.Topics.LATTICE_LINK_DEL] = linkDel

	// ------------------ Subscribe to New link topic --------------
	linkPut, err := wp.natsConnection.Subscribe(wp.Topics.LATTICE_LINK_PUT,
		func(m *nats.Msg) {
			link := InterfaceLinkDefinition{}
			err := json.Unmarshal(m.Data, &link)
			if err != nil {
				wp.Logger.Error(err, "failed to decode link")
				return
			}

			err = wp.putLink(link)
			if err != nil {
				// TODO(#10): handle this better?
				wp.Logger.Error(err, "newLinkFunc")
			}
		})
	if err != nil {
		wp.Logger.Error(err, "LINK_PUT")
		return err
	}
	wp.natsSubscriptions[wp.Topics.LATTICE_LINK_PUT] = linkPut

	// ------------------ Subscribe to Shutdown topic ------------------
	shutdown, err := wp.natsConnection.Subscribe(wp.Topics.LATTICE_SHUTDOWN,
		func(m *nats.Msg) {
			err := wp.shutdownFunc()
			if err != nil {
				// TODO(#10): handle this better?
				log.Print("ERROR: provider shutdown function failed: " + err.Error())
			}

			m.Respond([]byte("provider shutdown handled successfully"))
			wp.natsConnection.Flush()

			for _, s := range wp.natsSubscriptions {
				err := s.Drain()
				if err != nil {
					log.Print("ERROR: provider shutdown failed to drain subscription: " + err.Error())
				}
			}

			err = wp.natsConnection.Drain()
			if err != nil {
				log.Print("ERROR: provider shutdown failed to drain connection: " + err.Error())
			}

			wp.cancel()
		})
	if err != nil {
		wp.Logger.Error(err, "LATTICE_SHUTDOWN")
		return err
	}
	wp.natsSubscriptions[wp.Topics.LATTICE_SHUTDOWN] = shutdown
	return nil
}

func (wp *WasmcloudProvider) putLink(l InterfaceLinkDefinition) error {
	// Ignore duplicate links
	if wp.isLinked(l.SourceID, l.Target) {
		wp.Logger.Info("ignoring duplicate link", "link", l)
		return nil
	}

	wp.lock.Lock()
	defer wp.lock.Unlock()
	if l.SourceID == wp.Id {
		err := wp.putSourceLinkFunc(l)
		if err != nil {
			return err
		}

		wp.sourceLinks[l.Target] = l
	} else if l.Target == wp.Id {
		err := wp.putTargetLinkFunc(l)
		if err != nil {
			return err
		}

		wp.targetLinks[l.SourceID] = l
	} else {
		wp.Logger.Info("received link that isn't for this provider, ignoring", "link", l)
	}

	return nil
}

func (wp *WasmcloudProvider) deleteLink(l InterfaceLinkDefinition) error {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	if l.SourceID == wp.Id {
		err := wp.delSourceLinkFunc(l)
		if err != nil {
			return err
		}

		delete(wp.sourceLinks, l.Target)
	} else if l.Target == wp.Id {
		err := wp.delTargetLinkFunc(l)
		if err != nil {
			return err
		}
		delete(wp.targetLinks, l.SourceID)
	} else {
		wp.Logger.Info("received link delete that isn't for this provider, ignoring", "link", l)
	}

	return nil
}

func (wp *WasmcloudProvider) isLinked(sourceId string, target string) bool {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	if sourceId == wp.Id {
		_, exists := wp.sourceLinks[target]
		return exists
	} else if target == wp.Id {
		_, exists := wp.targetLinks[sourceId]
		return exists
	} else {
		return false
	}
}
