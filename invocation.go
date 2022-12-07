package provider

import (
	core "github.com/wasmcloud/interfaces/core/tinygo"
	msgpack "github.com/wasmcloud/tinygo-msgpack"
)

func EncodeInvocation(i core.Invocation) []byte {
	var sizeri msgpack.Sizer
	size_enci := &sizeri
	i.MEncode(size_enci)
	buf := make([]byte, sizeri.Len())
	encoderi := msgpack.NewEncoder(buf)
	enci := &encoderi
	i.MEncode(enci)
	return buf
}
