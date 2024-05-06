package provider

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"

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

	hostData   HostData
	Topics     Topics

	natsConnection    *nats.Conn
	natsSubscriptions map[string]*nats.Subscription

	healthMsgFunc func() string

	shutdownFunc func() error
	shutdown     chan struct{}

	// newLinkFunc func(core.LinkDefinition) error
	// delLinkFunc func(core.LinkDefinition) error

	lock  sync.Mutex
	// Links from the provider to other components, aka where the provider is the
	// source of the link. Indexed by the component ID of the target
	sourceLinks map[string]InterfaceLinkDefinition
    // Links from other components to the provider, aka where the provider is the
    // target of the link. Indexed by the component ID of the source
	targetLinks map[string]InterfaceLinkDefinition

	providerActionFunc func(ProviderAction) (*ProviderResponse, error)
}

func New(options ...ProviderOption) (*WasmcloudProvider, error) {
	logrusLog := logrus.New()
	logrusLog.SetFormatter(&logrus.JSONFormatter{})

	reader := bufio.NewReader(os.Stdin)
	// TODO: consider a better way to load?
	hostDataRaw, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	// hostDataDecoded, err := base64.StdEncoding.DecodeString(hostDataRaw)
	// if err != nil {
	// 	return nil, err
	// }

	hostData := HostData{}
	err = json.Unmarshal([]byte(hostDataRaw), &hostData)
	if err != nil {
		return nil, err
	}
	nc, err := nats.Connect(hostData.LatticeRPCURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	provider := &WasmcloudProvider{
		Id: hostData.ProviderKey,

		context: ctx,
		cancel:  cancel,
		Logger: logrusr.New(logrusLog).
			WithName(hostData.HostID),

		hostData:   hostData,
		Topics:     LatticeTopics(hostData),

		natsConnection:    nc,
		natsSubscriptions: map[string]*nats.Subscription{},

		healthMsgFunc: func() string { return "healthy" },

		shutdownFunc: func() error { return nil },
		shutdown:     make(chan struct{}),

		// newLinkFunc: func(core.LinkDefinition) error { return nil },
		// delLinkFunc: func(core.LinkDefinition) error { return nil },
		// newLink:     make(chan ActorConfig),
		// TODO: shouldn't be that big
		sourceLinks: make(map[string]InterfaceLinkDefinition, len(hostData.LinkDefinitions)),
		targetLinks: make(map[string]InterfaceLinkDefinition, len(hostData.LinkDefinitions)),

		providerActionFunc: func(a ProviderAction) (*ProviderResponse, error) {
			return &ProviderResponse{}, nil
		},
	}

	for _, opt := range options {
		err := opt(provider)
		if err != nil {
			return nil, err
		}
	}

	// TODO: start listening on existing links
	for _, link := range hostData.LinkDefinitions {
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
			// TODO: decode link
			// d := msgpack.NewDecoder(m.Data)
			// linkdef, err := core.MDecodeLinkDefinition(&d)
			// if err != nil {
			// 	wp.Logger.Error(err, "failed to decode link")
			// 	return
			// }

			// err = wp.delLinkFunc(linkdef)
			// if err != nil {
			// 	wp.Logger.Error(err, "failed to delete link")
			// 	return
			// }
		})
	if err != nil {
		wp.Logger.Error(err, "LINKDEF_DEL")
		return err
	}
	wp.natsSubscriptions[wp.Topics.LATTICE_LINK_DEL] = linkDel

	// ------------------ Subscribe to New link topic --------------
	linkPut, err := wp.natsConnection.Subscribe(wp.Topics.LATTICE_LINK_PUT,
		func(m *nats.Msg) {
			// TODO: decode link
			// d := msgpack.NewDecoder(m.Data)
			// linkdef, err := core.MDecodeLinkDefinition(&d)
			// if err != nil {
			// 	wp.Logger.Error(err, "failed to decode link")
			// 	return
			// }

			// err = wp.newLinkFunc(linkdef)
			// if err != nil {
			// 	// TODO: handle this better?
			// 	wp.Logger.Error(err, "newLinkFunc")
			// }
		})
	if err != nil {
		wp.Logger.Error(err, "LINKDEF_PUT")
		return err
	}
	wp.natsSubscriptions[wp.Topics.LATTICE_LINK_PUT] = linkPut

	// ------------------ Subscribe to Shutdown topic ------------------
	shutdown, err := wp.natsConnection.Subscribe(wp.Topics.LATTICE_SHUTDOWN,
		func(_ *nats.Msg) {
			err := wp.shutdownFunc()
			if err != nil {
				// TODO: handle this better?
				log.Print("ERROR: " + err.Error())
			}
		})
	if err != nil {
		wp.Logger.Error(err, "LATTICE_SHUTDOWN")
		return err
	}
	wp.natsSubscriptions[wp.Topics.LATTICE_SHUTDOWN] = shutdown
	return nil
}

func (wp *WasmcloudProvider) putLink(l InterfaceLinkDefinition) {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	// wp.links[l.ActorId] = l
}

func (wp *WasmcloudProvider) deleteLink(l InterfaceLinkDefinition) {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	// delete(wp.links, l.ActorId)
}

func (wp *WasmcloudProvider) isLinked(actorId string) bool {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	return false
}
