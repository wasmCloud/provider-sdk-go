package wasmcloud_core

import (
	"errors"

	msgpack "github.com/wasmcloud/tinygo-msgpack"
)

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

func (o *WasmcloudCoreTypesInvocation) MEncode_BC(encoder msgpack.Writer) error {
	encoder.WriteMapSize(9)
	encoder.WriteString("origin")
	o.Origin.MEncode_BC(encoder)
	encoder.WriteString("target")
	o.Target.MEncode_BC(encoder)
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
			b.MEncode_BC(encoder)
		}
	}

	return encoder.CheckError()
}

func (o *WasmcloudCoreTypesTuple2StringStringT) MEncode_BC(encoder msgpack.Writer) error {
	encoder.WriteString(o.F0)
	encoder.WriteString(o.F1)
	return encoder.CheckError()
}

func (o *WasmcloudCoreTypesWasmcloudEntity) MEncode_BC(encoder msgpack.Writer) error {
	switch o.Kind() {
	case WasmcloudCoreTypesWasmcloudEntityKindActor:
		encoder.WriteMapSize(3)
		encoder.WriteString("public_key")
		encoder.WriteString(o.GetActor())
		encoder.WriteString("link_name")
		encoder.WriteString("")
		encoder.WriteString("contract_id")
		encoder.WriteString("")
	case WasmcloudCoreTypesWasmcloudEntityKindProvider:
		encoder.WriteMapSize(3)
		encoder.WriteString("public_key")
		encoder.WriteString(o.GetProvider().PublicKey)
		encoder.WriteString("link_name")
		encoder.WriteString(o.GetProvider().LinkName)
		encoder.WriteString("contract_id")
		encoder.WriteString(o.GetProvider().ContractId)
	default:
		return errors.New("invalid kind of wasmcloud entitiy")

	}

	return encoder.CheckError()
}

func (o *WasmcloudCoreTypesWasmcloudEntity) MEncode(encoder msgpack.Writer) error {
	switch o.Kind() {
	case WasmcloudCoreTypesWasmcloudEntityKindActor:
		actor := o.GetActor()
		encoder.WriteMapSize(1)
		encoder.WriteString("public_key")
		encoder.WriteString(actor)
	case WasmcloudCoreTypesWasmcloudEntityKindProvider:
		provider := o.GetProvider()
		encoder.WriteMapSize(3)
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

func (o *WasmcloudCoreTypesLinkDefinition) MEncode(encoder msgpack.Writer) error {
	encoder.WriteMapSize(5)
	encoder.WriteString("actor_id")
	encoder.WriteString(o.ActorId)
	encoder.WriteString("provider_id")
	encoder.WriteString(o.ProviderId)
	encoder.WriteString("link_name")
	encoder.WriteString(o.LinkName)
	encoder.WriteString("contract_id")
	encoder.WriteString(o.ContractId)
	encoder.WriteString("values")
	if o.Values.IsNone() {
		encoder.WriteNil()
	} else {
		o.Values.MEncode(encoder)
	}
	return encoder.CheckError()
}

func (o WasmcloudCoreTypesCapabilityContractId) MEncode(encoder msgpack.Writer) error {
	encoder.WriteString(string(o))
	return encoder.CheckError()
}

func (o *Option[T]) MEncode(encoder msgpack.Writer) error {
	if o.IsSome() {
		val := o.Unwrap()
		switch tVal := any(val).(type) {
		case []WasmcloudCoreTypesTuple2StringStringT:
			encoder.WriteArraySize(uint32(len(tVal)))
			for _, b := range tVal {
				b.MEncode(encoder)
			}
		case string:
			encoder.WriteString(tVal)
		default:
			return errors.New("invalid option type")
		}
	}
	return encoder.CheckError()
}

func (o *WasmcloudCoreTypesInvocationResponse) MEncode(encoder msgpack.Writer) error {
	encoder.WriteMapSize(4)
	encoder.WriteString("msg")
	encoder.WriteByteArray(o.Msg)
	encoder.WriteString("invocation_id")
	encoder.WriteString(o.InvocationId)
	encoder.WriteString("error")
	if o.Error.IsNone() {
		encoder.WriteNil()
	} else {
		o.Error.MEncode(encoder)
	}
	encoder.WriteString("content_length")
	encoder.WriteUint64(o.ContentLength)

	return encoder.CheckError()
}
