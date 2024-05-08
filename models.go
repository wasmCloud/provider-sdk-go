package provider

import (
	"fmt"
)

type Topics struct {
	LATTICE_LINK_GET string
	LATTICE_LINK_DEL string
	LATTICE_LINK_PUT string
	LATTICE_SHUTDOWN string
	LATTICE_HEALTH   string
}

func LatticeTopics(h HostData) Topics {
	return Topics{
		LATTICE_LINK_GET: fmt.Sprintf("wasmbus.rpc.%s.%s.linkdefs.get", h.LatticeRPCPrefix, h.ProviderKey),
		LATTICE_LINK_DEL: fmt.Sprintf("wasmbus.rpc.%s.%s.linkdefs.del", h.LatticeRPCPrefix, h.ProviderKey),
		LATTICE_LINK_PUT: fmt.Sprintf("wasmbus.rpc.%s.%s.linkdefs.put", h.LatticeRPCPrefix, h.ProviderKey),
		LATTICE_HEALTH:   fmt.Sprintf("wasmbus.rpc.%s.%s.health", h.LatticeRPCPrefix, h.ProviderKey),
		LATTICE_SHUTDOWN: fmt.Sprintf("wasmbus.rpc.%s.%s.default.shutdown", h.LatticeRPCPrefix, h.ProviderKey),
	}
}
