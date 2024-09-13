package wrpchttp

import (
	"errors"

	wrpcnats "wrpc.io/go/nats"
)

var (
	ErrNoTarget = errors.New("no target")
	ErrRPC      = errors.New("rpc error")
)

type NatsClientCreator interface {
	OutgoingRpcClient(target string) *wrpcnats.Client
}
