package provider

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"

	core "github.com/wasmcloud/interfaces/core/tinygo"
	msgpack "github.com/wasmcloud/tinygo-msgpack"
)

type WasmcloudProvider struct {
	Id string

	context context.Context
	cancel  context.CancelFunc
	Logger  logr.Logger

	hostData   core.HostData
	contractId string
	Topics     Topics

	natsConnection    *nats.Conn
	natsSubscriptions map[string]*nats.Subscription

	healthMsgFunc func() string

	shutdownFunc func() error
	shutdown     chan struct{}

	newLinkFunc func(core.LinkDefinition) error
	delLinkFunc func(core.LinkDefinition) error

	lock  sync.Mutex
	links map[string]core.LinkDefinition

	providerActionFunc func(ProviderAction) (*ProviderResponse, error)
}

func New(contract string, options ...ProviderOption) (*WasmcloudProvider, error) {
	logrusLog := logrus.New()
	logrusLog.SetFormatter(&logrus.JSONFormatter{})

	reader := bufio.NewReader(os.Stdin)
	hostDataRaw, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	hostDataDecoded, err := base64.StdEncoding.DecodeString(hostDataRaw)
	if err != nil {
		return nil, err
	}

	hostData := core.HostData{}
	err = json.Unmarshal([]byte(hostDataDecoded), &hostData)
	if err != nil {
		return nil, err
	}
	nc, err := nats.Connect(hostData.LatticeRpcUrl)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	provider := &WasmcloudProvider{
		Id: hostData.ProviderKey,

		context: ctx,
		cancel:  cancel,
		Logger: logrusr.New(logrusLog).
			WithName(hostData.HostId),

		hostData:   hostData,
		contractId: contract,
		Topics:     LatticeTopics(hostData),

		natsConnection:    nc,
		natsSubscriptions: map[string]*nats.Subscription{},

		healthMsgFunc: func() string { return "healthy" },

		shutdownFunc: func() error { return nil },
		shutdown:     make(chan struct{}),

		newLinkFunc: func(core.LinkDefinition) error { return nil },
		delLinkFunc: func(core.LinkDefinition) error { return nil },
		// newLink:     make(chan ActorConfig),
		links: make(map[string]core.LinkDefinition, len(hostData.LinkDefinitions)),

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
		if link.ProviderId == provider.Id {
			provider.newLinkFunc(link)
		}
	}

	return provider, nil
}

func (wp *WasmcloudProvider) HostData() core.HostData {
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

func (wp *WasmcloudProvider) listenForActor(actorID string) {
	subj := fmt.Sprintf("wasmbus.rpc.%s.%s.%s",
		wp.hostData.LatticeRpcPrefix,
		wp.hostData.ProviderKey,
		wp.hostData.LinkName,
	)

	actorsub, err := wp.natsConnection.Subscribe(subj,
		func(m *nats.Msg) {
			d := msgpack.NewDecoder(m.Data)
			i, err := core.MDecodeInvocation(&d)
			if err != nil {
				return
			}

			if err := wp.validateProviderInvocation(i); err != nil {
				wp.Logger.Error(err, "validate provider invocation failed")
				return
			}

			payload := ProviderAction{
				Operation: i.Operation,
				Msg:       i.Msg,
				FromActor: actorID,
			}

			// TODO: need to set default
			resp, err := wp.providerActionFunc(payload)
			if err != nil {
				// TODO: what to do with this error
				return
			}

			ir := core.InvocationResponse{
				Msg:           resp.Msg,
				Error:         resp.Error,
				InvocationId:  i.Id,
				ContentLength: uint64(len(resp.Msg)),
			}

			buf := MEncode(&ir)

			wp.natsConnection.Publish(m.Reply, buf)
		})

	if err != nil {
		wp.Logger.Error(err, "ACTOR_SUB")
		return
	}

	wp.natsSubscriptions[actorID] = actorsub
}

func (wp *WasmcloudProvider) validateProviderInvocation(invocation core.Invocation) error {
	// todo validate claims issuer is included in cluster issuers

	if invocation.Target.PublicKey != wp.hostData.ProviderKey {
		return fmt.Errorf("target key mismatch: %s != %s", invocation.Target.PublicKey, wp.hostData.HostId)
	}

	if !wp.isLinked(invocation.Origin.PublicKey) {
		return fmt.Errorf("unlinked actor: %s", invocation.Origin.PublicKey)
	}
	return nil
}

func (wp *WasmcloudProvider) subToNats() error {
	// ------------------ Subscribe to Health topic --------------------
	health, err := wp.natsConnection.Subscribe(wp.Topics.LATTICE_HEALTH,
		func(m *nats.Msg) {
			hc := core.HealthCheckResponse{
				Healthy: true,
				Message: wp.healthMsgFunc(),
			}

			buf := MEncode(&hc)

			wp.natsConnection.Publish(m.Reply, buf)
		})
	if err != nil {
		wp.Logger.Error(err, "LATTICE_HEALTH")
		return err
	}
	wp.natsSubscriptions[wp.Topics.LATTICE_HEALTH] = health

	// ------------------ Subscribe to Delete link topic --------------
	linkDel, err := wp.natsConnection.Subscribe(wp.Topics.LATTICE_LINKDEF_DEL,
		func(m *nats.Msg) {
			d := msgpack.NewDecoder(m.Data)
			linkdef, err := core.MDecodeLinkDefinition(&d)
			if err != nil {
				wp.Logger.Error(err, "failed to decode link")
				return
			}

			err = wp.delLinkFunc(linkdef)
			if err != nil {
				wp.Logger.Error(err, "failed to delete link")
				return
			}
		})
	if err != nil {
		wp.Logger.Error(err, "LINKDEF_DEL")
		return err
	}
	wp.natsSubscriptions[wp.Topics.LATTICE_LINKDEF_DEL] = linkDel

	// ------------------ Subscribe to New link topic --------------
	linkPut, err := wp.natsConnection.Subscribe(wp.Topics.LATTICE_LINKDEF_PUT,
		func(m *nats.Msg) {
			d := msgpack.NewDecoder(m.Data)
			linkdef, err := core.MDecodeLinkDefinition(&d)
			if err != nil {
				wp.Logger.Error(err, "failed to decode link")
				return
			}

			err = wp.newLinkFunc(linkdef)
			if err != nil {
				// TODO: handle this better?
				wp.Logger.Error(err, "newLinkFunc")
			}
		})
	if err != nil {
		wp.Logger.Error(err, "LINKDEF_PUT")
		return err
	}
	wp.natsSubscriptions[wp.Topics.LATTICE_LINKDEF_PUT] = linkPut

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

func (wp *WasmcloudProvider) ToActor(actorID string, msg []byte, op string) ([]byte, error) {
	guid := uuid.New().String()

	i := core.Invocation{
		Origin: core.WasmCloudEntity{
			PublicKey:  wp.hostData.ProviderKey,
			LinkName:   wp.hostData.LinkName,
			ContractId: core.CapabilityContractId(wp.contractId),
		},
		Target: core.WasmCloudEntity{
			PublicKey:  actorID,
			LinkName:   wp.hostData.LinkName,
			ContractId: core.CapabilityContractId(wp.contractId),
		},
		Operation:     op,
		Msg:           msg,
		Id:            guid,
		HostId:        wp.hostData.HostId,
		ContentLength: uint64(len([]byte(msg))),
	}

	err := EncodeClaims(&i, wp.hostData, guid)
	if err != nil {
		wp.Logger.Error(err, "Failed to encode claims")
		return nil, err
	}

	natsBody := MEncode(&i)

	// NC Request
	subj := fmt.Sprintf("wasmbus.rpc.%s.%s", wp.hostData.LatticeRpcPrefix, actorID)
	ir, err := wp.natsConnection.Request(subj, natsBody, 2*time.Second)
	if err != nil {
		wp.Logger.Error(err, "NATs request failed")
		return nil, err
	}

	d := msgpack.NewDecoder(ir.Data)
	resp, err := core.MDecodeInvocationResponse(&d)
	if err != nil {
		wp.Logger.Error(err, "Failed to decode invocation response")
		return nil, err
	}

	return resp.Msg, nil
}

func (wp *WasmcloudProvider) putLink(l core.LinkDefinition) {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	wp.links[l.ActorId] = l
}

func (wp *WasmcloudProvider) deleteLink(l core.LinkDefinition) {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	delete(wp.links, l.ActorId)
}

func (wp *WasmcloudProvider) isLinked(actorId string) bool {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	_, exist := wp.links[actorId]
	return exist
}
