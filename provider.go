package provider

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	context context.Context
	cancel  context.CancelFunc
	Logger  logr.Logger

	hostData   core.HostData
	contractId string
	Topics     Topics

	natsConnection    *nats.Conn
	natsSubscriptions []*nats.Subscription

	healthMsgFunc func() string

	shutdownFunc func() error
	shutdown     chan struct{}

	newLinkFunc func(ActorConfig) error
	newLink     chan ActorConfig
	links       []ActorConfig

	providerActionFunc func(ProviderAction) (*ProviderResponse, error)
}

func New(contract string, options ...func(*WasmcloudProvider) error) (*WasmcloudProvider, error) {
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
		context: ctx,
		cancel:  cancel,
		Logger: logrusr.New(logrusLog).
			WithName(hostData.HostId),

		hostData:   hostData,
		contractId: contract,
		Topics:     LatticeTopics(hostData),

		natsConnection:    nc,
		natsSubscriptions: []*nats.Subscription{},

		healthMsgFunc: func() string { return "healthy" },

		shutdownFunc: func() error { return nil },
		shutdown:     make(chan struct{}),

		newLinkFunc: func(ActorConfig) error { return nil },
		newLink:     make(chan ActorConfig),
		links:       []ActorConfig{},

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

	return provider, nil
}

func (wp WasmcloudProvider) HostData() core.HostData {
	return wp.hostData
}

func (wp WasmcloudProvider) Start() error {
	err := wp.subToNats()
	if err != nil {
		return err
	}

startloop:
	for {
		select {
		case actorData := <-wp.newLink:
			err := wp.newLinkFunc(actorData)
			if err != nil {
				// TODO: handle this better?
				log.Print("ERROR: " + err.Error())
			}
		case <-wp.shutdown:
			err := wp.shutdownFunc()
			if err != nil {
				// TODO: handle this better?
				log.Print("ERROR: " + err.Error())
			}
		case <-wp.context.Done():
			break startloop
		}
	}

	return nil
}

func (p *WasmcloudProvider) listenForActor(actorID string) {
	subj := fmt.Sprintf("wasmbus.rpc.%s.%s.%s",
		p.hostData.LinkName,
		p.hostData.ProviderKey,
		p.hostData.LinkName,
	)

	p.natsConnection.Subscribe(subj,
		func(m *nats.Msg) {
			d := msgpack.NewDecoder(m.Data)
			i, err := core.MDecodeInvocation(&d)
			if err != nil {
				return
			}

			payload := ProviderAction{
				Operation: i.Operation,
				Msg:       i.Msg,
				FromActor: actorID,
			}

			// TODO: need to set default
			resp, err := p.providerActionFunc(payload)
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

			var sizer msgpack.Sizer
			size_enc := &sizer
			ir.MEncode(size_enc)
			buf := make([]byte, sizer.Len())
			encoder := msgpack.NewEncoder(buf)
			enc := &encoder
			ir.MEncode(enc)

			p.natsConnection.Publish(m.Reply, buf)
		})
}

func (p *WasmcloudProvider) subToNats() error {
	linkGet, err := p.natsConnection.QueueSubscribe(
		p.Topics.LATTICE_LINKDEF_GET,
		p.Topics.LATTICE_LINKDEF_GET,
		func(m *nats.Msg) {
			var sizer msgpack.Sizer
			size_enc := &sizer
			p.hostData.LinkDefinitions.MEncode(size_enc)
			buf := make([]byte, sizer.Len())
			encoder := msgpack.NewEncoder(buf)
			enc := &encoder
			p.hostData.LinkDefinitions.MEncode(enc)
			p.natsConnection.Publish(m.Reply, buf)
		})
	if err != nil {
		p.Logger.Error(err, "LINKDEF_GET: "+p.Topics.LATTICE_LINKDEF_GET)
		return err
	}

	p.natsSubscriptions = append(p.natsSubscriptions, linkGet)

	health, err := p.natsConnection.Subscribe(p.Topics.LATTICE_HEALTH,
		func(m *nats.Msg) {
			hc := core.HealthCheckResponse{
				Healthy: true,
				Message: p.healthMsgFunc(),
			}

			var sizer msgpack.Sizer
			size_enc := &sizer
			hc.MEncode(size_enc)
			buf := make([]byte, sizer.Len())
			encoder := msgpack.NewEncoder(buf)
			enc := &encoder
			hc.MEncode(enc)

			p.natsConnection.Publish(m.Reply, buf)
		})
	if err != nil {
		p.Logger.Error(err, "LATTICE_HEALTH")
		return err
	}

	p.natsSubscriptions = append(p.natsSubscriptions, health)

	linkDel, err := p.natsConnection.Subscribe(p.Topics.LATTICE_LINKDEF_DEL,
		func(m *nats.Msg) {
			d := msgpack.NewDecoder(m.Data)

			// TODO: Handle a deleted link correctly
			_, err := core.MDecodeLinkDefinition(&d)
			if err != nil {
				return
			}
		})
	if err != nil {
		p.Logger.Error(err, "LINKDEF_DEL")
		return err
	}
	p.natsSubscriptions = append(p.natsSubscriptions, linkDel)

	linkPut, err := p.natsConnection.Subscribe(p.Topics.LATTICE_LINKDEF_PUT,
		func(m *nats.Msg) {
			d := msgpack.NewDecoder(m.Data)
			linkdef, err := core.MDecodeLinkDefinition(&d)
			if err != nil {
				return
			}

			p.newLink <- ActorConfig{linkdef.ActorId, linkdef.Values}
		})
	if err != nil {
		p.Logger.Error(err, "LINKDEF_PUT")
		return err
	}
	p.natsSubscriptions = append(p.natsSubscriptions, linkPut)

	shutdown, err := p.natsConnection.Subscribe(p.Topics.LATTICE_SHUTDOWN,
		func(_ *nats.Msg) {
			p.shutdown <- struct{}{}
		})
	if err != nil {
		p.Logger.Error(err, "LATTICE_SHUTDOWN")
		return err
	}
	p.natsSubscriptions = append(p.natsSubscriptions, shutdown)
	return nil
}

func (wp WasmcloudProvider) SendDownLattice(actorID string, msg []byte, op string) ([]byte, error) {
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
		return nil, err
	}

	natsBody := EncodeInvocation(i)

	// NC Request
	subj := fmt.Sprintf("wasmbus.rpc.%s.%s", wp.hostData.LatticeRpcPrefix, actorID)
	wp.Logger.Info("Encoded invocation sent",
		map[string]interface{}{
			"subj": subj,
			"op":   op,
			"msg":  msg,
		},
	)

	ir, err := wp.natsConnection.Request(subj, natsBody, 2*time.Second)
	if err != nil {
		return nil, err
	}

	// d := cbor.NewDecoder(ir.Data)
	// resp, err := core.CDecodeInvocationResponse(&d)
	d := msgpack.NewDecoder(ir.Data)
	resp, err := core.MDecodeInvocationResponse(&d)
	wp.Logger.Error(err, "SDL")

	return resp.Msg, nil
}
