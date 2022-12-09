package provider

import (
	"fmt"

	"github.com/golang-jwt/jwt"
	core "github.com/wasmcloud/interfaces/core/tinygo"
)

type Claims struct {
	jwt.StandardClaims
	ID     string `json:"jti"`
	Wascap Wascap `json:"wascap"`
}

type Wascap struct {
	TargetURL string `json:"target_url"`
	OriginURL string `json:"origin_url"`
	Hash      string `json:"hash"`
}

type ProviderResponse struct {
	Msg   []byte `msgpack:"msg,omitempty"`
	Error string `msgpack:"error,omitempty"`
}

type Topics struct {
	LATTICE_LINKDEF_GET string
	LATTICE_LINKDEF_DEL string
	LATTICE_LINKDEF_PUT string
	LATTICE_SHUTDOWN    string
	LATTICE_HEALTH      string
}

func LatticeTopics(h core.HostData) Topics {
	return Topics{
		LATTICE_LINKDEF_GET: fmt.Sprintf("wasmbus.rpc.%s.%s.%s.linkdefs.get", h.LatticeRpcPrefix, h.ProviderKey, h.LinkName),
		LATTICE_LINKDEF_DEL: fmt.Sprintf("wasmbus.rpc.%s.%s.%s.linkdefs.del", h.LatticeRpcPrefix, h.ProviderKey, h.LinkName),
		LATTICE_LINKDEF_PUT: fmt.Sprintf("wasmbus.rpc.%s.%s.%s.linkdefs.put", h.LatticeRpcPrefix, h.ProviderKey, h.LinkName),
		LATTICE_SHUTDOWN:    fmt.Sprintf("wasmbus.rpc.%s.%s.%s.shutdown", h.LatticeRpcPrefix, h.ProviderKey, h.LinkName),
		LATTICE_HEALTH:      fmt.Sprintf("wasmbus.rpc.%s.%s.%s.health", h.LatticeRpcPrefix, h.ProviderKey, h.LinkName),
	}
}

type ProviderAction struct {
	Operation string
	Msg       []byte
	FromActor string
	// Respond   chan ProviderResponse
}

type LinkDefinition struct {
	ActorID    string            `msgpack:"actor_id"`
	ProviderID string            `msgpack:"provider_id"`
	LinkName   string            `msgpack:"link_name"`
	ContractID string            `msgpack:"contract_id"`
	Values     map[string]string `msgpack:"values"`
}
