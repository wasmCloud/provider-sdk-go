package wasmcloud_core

import (
	msgpack "github.com/wasmcloud/tinygo-msgpack"
)

func (o *WasmcloudCoreTypesInvocationResponse) MEncode(encoder msgpack.Writer) error {
	encoder.WriteMapSize(4)
	encoder.WriteString("msg")
	encoder.WriteByteArray(o.Msg)
	encoder.WriteString("invocation_id")
	encoder.WriteString(o.InvocationId)
	encoder.WriteString("error")
	if o.Error.IsSome() {
		encoder.WriteString(o.Error.Unwrap())
	} else {
		encoder.WriteNil()
	}
	encoder.WriteString("content_length")
	encoder.WriteUint64(o.ContentLength)

	return encoder.CheckError()
}

func (o *WasmcloudCoreTypesHealthCheckResponse) MEncode(encoder msgpack.Writer) error {
	encoder.WriteMapSize(2)
	encoder.WriteString("healthy")
	encoder.WriteBool(o.Healthy)
	encoder.WriteString("message")
	encoder.WriteString(o.Message)

	return encoder.CheckError()
}

func (o *WasmcloudCoreTypesInvocation) MEncode(encoder msgpack.Writer) error {
	encoder.WriteMapSize(9)
	encoder.WriteString("origin")
	o.Origin.MEncode(encoder)
	encoder.WriteString("target")
	o.Target.MEncode(encoder)
	encoder.WriteString("operation")
	encoder.WriteString(o.Operation)
	encoder.WriteString("msg")
	encoder.WriteByteArray(o.Msg)
	encoder.WriteString("id")
	encoder.WriteString(o.Id)
	encoder.WriteString("encoded_claims")
	encoder.WriteString(o.EncodedClaims)
	encoder.WriteString("host_id")
	encoder.WriteString(o.SourceHostId)
	encoder.WriteString("content_length")
	encoder.WriteUint64(o.ContentLength)
	encoder.WriteString("traceContext")
	if o.TraceContext == nil {
		encoder.WriteNil()
	} else {
		o.TraceContext.MEncode(encoder)
	}

	return encoder.CheckError()
}
