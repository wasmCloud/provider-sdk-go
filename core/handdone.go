package wasmcloud_core

import (
	"errors"

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
	err := o.Origin.MEncode(encoder)
	if err != nil {
		return err
	}
	encoder.WriteString("target")
	err = o.Target.MEncode(encoder)
	if err != nil {
		return err
	}
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
		encoder.WriteArraySize(uint32(len(o.TraceContext)))
		for _, b := range o.TraceContext {
			b.MEncode(encoder)
		}
	}

	return encoder.CheckError()
}

func (o *WasmcloudCoreTypesWasmcloudEntity) MEncode(encoder msgpack.Writer) error {
	switch o.Kind() {
	case 0: //actor
		actor := o.GetActor()
		encoder.WriteMapSize(2)
		encoder.WriteString("entity_type")
		encoder.WriteInt8(0)
		encoder.WriteString("public_key")
		encoder.WriteString(actor)
	case 1: //provider
		provider := o.GetProvider()
		encoder.WriteMapSize(4)
		encoder.WriteString("entity_type")
		encoder.WriteInt8(1)
		encoder.WriteString("public_key")
		encoder.WriteString(provider.PublicKey)
		encoder.WriteString("link_name")
		encoder.WriteString(provider.LinkName)
		encoder.WriteString("contract_id")
		encoder.WriteString(provider.ContractId)
	default:
		return errors.New("invalid kind of wasmcloud entitiy")
	}

	return encoder.CheckError()
}

func (o *WasmcloudCoreTypesTuple2StringStringT) MEncode(encoder msgpack.Writer) error {
	encoder.WriteMapSize(2)
	encoder.WriteString("f0")
	encoder.WriteString(o.F0)
	encoder.WriteString("f1")
	encoder.WriteString(o.F1)
	return encoder.CheckError()
}