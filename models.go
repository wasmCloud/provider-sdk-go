package provider

import (
	"fmt"

	"github.com/golang-jwt/jwt"
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
	LATTICE_LINK_GET string
	LATTICE_LINK_DEL string
	LATTICE_LINK_PUT string
	LATTICE_SHUTDOWN    string
	LATTICE_HEALTH      string
}

func LatticeTopics(h HostData) Topics {
	return Topics{
		LATTICE_LINK_GET: fmt.Sprintf("wasmbus.rpc.%s.%s.%s.linkdefs.get", h.LatticeRPCPrefix, h.ProviderKey, h.LinkName),
		LATTICE_LINK_DEL: fmt.Sprintf("wasmbus.rpc.%s.%s.%s.linkdefs.del", h.LatticeRPCPrefix, h.ProviderKey, h.LinkName),
		LATTICE_LINK_PUT: fmt.Sprintf("wasmbus.rpc.%s.%s.%s.linkdefs.put", h.LatticeRPCPrefix, h.ProviderKey, h.LinkName),
		LATTICE_SHUTDOWN:    fmt.Sprintf("wasmbus.rpc.%s.%s.%s.shutdown", h.LatticeRPCPrefix, h.ProviderKey, h.LinkName),
		LATTICE_HEALTH:      fmt.Sprintf("wasmbus.rpc.%s.%s.%s.health", h.LatticeRPCPrefix, h.ProviderKey, h.LinkName),
	}
}

type ProviderAction struct {
	Operation string
	Msg       []byte
	FromActor string
	// Respond   chan ProviderResponse
}
