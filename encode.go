package provider

import msgpack "github.com/wasmcloud/tinygo-msgpack"

type MEncodable interface {
	MEncode(encoder msgpack.Writer) error
}

func MEncode(v MEncodable) []byte {
	var sizer msgpack.Sizer
	v.MEncode(&sizer)
	buf := make([]byte, sizer.Len())
	encoder := msgpack.NewEncoder(buf)
	v.MEncode(&encoder)
	return buf
}
