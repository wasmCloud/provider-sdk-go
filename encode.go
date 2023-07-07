package provider

import (
	wasmcloud_core "github.com/wasmCloud/provider-sdk-go/core"
	msgpack "github.com/wasmcloud/tinygo-msgpack"
)

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

func MDecodeInvocation(d *msgpack.Decoder) (wasmcloud_core.WasmcloudCoreTypesInvocation, error) {
	var val wasmcloud_core.WasmcloudCoreTypesInvocation
	isNil, err := d.IsNextNil()
	if err != nil || isNil {
		return val, err
	}
	size, err := d.ReadMapSize()
	if err != nil {
		return val, err
	}
	for i := uint32(0); i < size; i++ {
		field, err := d.ReadString()
		if err != nil {
			return val, err
		}
		switch field {
		case "origin":
			val.Origin, err = MDecodeWasmCloudEntity(d)
		case "target":
			val.Target, err = MDecodeWasmCloudEntity(d)
		case "operation":
			val.Operation, err = d.ReadString()
		case "msg":
			val.Msg, err = d.ReadByteArray()
		case "id":
			val.Id, err = d.ReadString()
		case "encoded_claims":
			val.EncodedClaims, err = d.ReadString()
		case "host_id":
			val.SourceHostId, err = d.ReadString()
		case "content_length":
			val.ContentLength, err = d.ReadUint64()
		case "traceContext":
			val.TraceContext, err = MDecodeTraceContext(d)
		default:
			err = d.Skip()
		}
		if err != nil {
			return val, err
		}
	}
	return val, nil
}

func MDecodeLinkDefinition(d *msgpack.Decoder) (wasmcloud_core.WasmcloudCoreTypesLinkDefinition, error) {
	var val wasmcloud_core.WasmcloudCoreTypesLinkDefinition
	isNil, err := d.IsNextNil()
	if err != nil || isNil {
		return val, err
	}
	size, err := d.ReadMapSize()
	if err != nil {
		return val, err
	}
	for i := uint32(0); i < size; i++ {
		field, err := d.ReadString()
		if err != nil {
			return val, err
		}
		switch field {
		case "actor_id":
			val.ActorId, err = d.ReadString()
		case "provider_id":
			val.ProviderId, err = d.ReadString()
		case "link_name":
			val.LinkName, err = d.ReadString()
		case "contract_id":
			val.ContractId, err = d.ReadString()
		case "values":
			val.Values, err = MDecodeLinkSettings(d)
		default:
			err = d.Skip()
		}
		if err != nil {
			return val, err
		}
	}
	return val, nil
}

func MDecodeLinkSettings(d *msgpack.Decoder) (wasmcloud_core.WasmcloudCoreTypesLinkSettings, error) {
	isNil, err := d.IsNextNil()
	if err != nil || isNil {
		return make(map[string]string, 0), err
	}
	size, err := d.ReadMapSize()
	if err != nil {
		return make(map[string]string, 0), err
	}
	val := make(map[string]string, size)
	for i := uint32(0); i < size; i++ {
		k, _ := d.ReadString()
		v, err := d.ReadString()
		if err != nil {
			return val, err
		}
		val[k] = v
	}
	return val, nil
}

func MDecodeInvocationResponse(d *msgpack.Decoder) (wasmcloud_core.WasmcloudCoreTypesInvocationResponse, error) {
	var val wasmcloud_core.WasmcloudCoreTypesInvocationResponse
	isNil, err := d.IsNextNil()
	if err != nil || isNil {
		return val, err
	}
	size, err := d.ReadMapSize()
	if err != nil {
		return val, err
	}
	for i := uint32(0); i < size; i++ {
		field, err := d.ReadString()
		if err != nil {
			return val, err
		}
		switch field {
		case "msg":
			val.Msg, err = d.ReadByteArray()
		case "invocation_id":
			val.InvocationId, err = d.ReadString()
		case "error":
			val.Error, err = d.ReadString()
		case "content_length":
			val.ContentLength, err = d.ReadUint64()
		default:
			err = d.Skip()
		}
		if err != nil {
			return val, err
		}
	}
	return val, nil
}

func MDecodeWasmCloudEntity(d *msgpack.Decoder) (wasmcloud_core.WasmcloudCoreTypesWasmcloudEntity, error) {
	var val wasmcloud_core.WasmcloudCoreTypesWasmcloudEntity
	isNil, err := d.IsNextNil()
	if err != nil || isNil {
		return val, err
	}
	size, err := d.ReadMapSize()
	if err != nil {
		return val, err
	}
	for i := uint32(0); i < size; i++ {
		field, err := d.ReadString()
		if err != nil {
			return val, err
		}
		switch field {
		case "public_key":
			val.PublicKey, err = d.ReadString()
		case "link_name":
			val.LinkName, err = d.ReadString()
		case "contract_id":
			val.ContractId, err = MDecodeCapabilityContractId(d)
		default:
			err = d.Skip()
		}
		if err != nil {
			return val, err
		}
	}
	return val, nil
}

func MDecodeTraceContext(d *msgpack.Decoder) (wasmcloud_core.WasmcloudCoreTypesTraceContext, error) {
	isNil, err := d.IsNextNil()
	if err != nil || isNil {
		return make(map[string]string, 0), err
	}
	size, err := d.ReadMapSize()
	if err != nil {
		return make(map[string]string, 0), err
	}
	val := make(map[string]string, size)
	for i := uint32(0); i < size; i++ {
		k, _ := d.ReadString()
		v, err := d.ReadString()
		if err != nil {
			return val, err
		}
		val[k] = v
	}
	return val, nil
}

func MDecodeCapabilityContractId(d *msgpack.Decoder) (wasmcloud_core.WasmcloudCoreTypesCapabilityContractId, error) {
	val, err := d.ReadString()
	if err != nil {
		return "", err
	}
	return wasmcloud_core.WasmcloudCoreTypesCapabilityContractId(val), nil
}

func MDecodeLinkSettings(d *msgpack.Decoder) (wasmcloud_core.WasmcloudCoreTypesLinkSettings, error) {
	isNil, err := d.IsNextNil()
	if err != nil || isNil {
		return make(map[string]string, 0), err
	}
	size, err := d.ReadMapSize()
	if err != nil {
		return make(map[string]string, 0), err
	}
	val := make(map[string]string, size)
	for i := uint32(0); i < size; i++ {
		k, _ := d.ReadString()
		v, err := d.ReadString()
		if err != nil {
			return val, err
		}
		val[k] = v
	}
	return val, nil
}
