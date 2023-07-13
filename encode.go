package provider

import (
	"errors"

	wasmcloud_core "github.com/wasmCloud/provider-sdk-go/core"
	msgpack "github.com/wasmcloud/tinygo-msgpack"
)

type MEncodable interface {
	MEncode(encoder msgpack.Writer) error
}

type MEncodable_BC interface {
	MEncode_BC(encoder msgpack.Writer) error
}

func MEncode(v MEncodable) []byte {
	var sizer msgpack.Sizer
	v.MEncode(&sizer)
	buf := make([]byte, sizer.Len())
	encoder := msgpack.NewEncoder(buf)
	v.MEncode(&encoder)
	return buf
}

func MEncode_BC(v MEncodable_BC) []byte {
	var sizer msgpack.Sizer
	v.MEncode_BC(&sizer)
	buf := make([]byte, sizer.Len())
	encoder := msgpack.NewEncoder(buf)
	v.MEncode_BC(&encoder)
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
			// TODO: values are an option<tuple<list>> in wit.  Needs a BC decoder
			// that I haven't written yet
			//
		// case "values":
		// 	isNone, err := d.IsNextNil() // means Option == None
		// 	if err != nil {
		// 		return val, err
		// 	}
		//
		// 	if !isNone {
		// 		var v []wasmcloud_core.WasmcloudCoreTypesTuple2StringStringT
		// 		opt := wasmcloud_core.Option[[]wasmcloud_core.WasmcloudCoreTypesTuple2StringStringT]{}
		// 		v, err = MDecodeLinkSettings(d)
		// 		if err != nil {
		// 			return val, err
		// 		}
		// 		opt.Set(v)
		// 		val.Values = opt
		// 	}
		default:
			err = d.Skip()
		}
		if err != nil {
			return val, err
		}
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
			if err != nil {
				val.Msg = nil
				err = nil
			}
		case "invocation_id":
			val.InvocationId, err = d.ReadString()
		case "error":
			isNone, err := d.IsNextNil() // means Option == None
			if err != nil {
				return val, err
			}

			if !isNone {
				var v string
				opt := wasmcloud_core.Option[string]{}
				v, err = d.ReadString()
				if err != nil {
					return val, err
				}
				opt.Set(v)
				val.Error = opt
			}
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

	field, err := d.ReadString() //entity_type field
	if err != nil {
		return val, err
	}

	var eType int8
	if field == "entity_type" {
		eType, err = d.ReadInt8()
		if err != nil {
			return val, err
		}
	} else {
		return val, errors.New("unexpected field in entity_type location")
	}

	switch eType {
	case 0: // actor_id
		for i := uint32(1); i < size; i++ {
			field, err := d.ReadString() // public_key field
			if err != nil {
				return val, err
			}

			switch field {
			case "public_key":
				ps, err := d.ReadString()
				if err != nil {
					return val, err
				}
				val.SetActor(ps)
			default:
				d.Skip()
			}

		}

	case 1: // provider
		weid := wasmcloud_core.WasmcloudCoreTypesProviderIdentifier{}

		for i := uint32(1); i < size; i++ {
			field, err := d.ReadString() // public_key field
			if err != nil {
				return val, err
			}

			switch field {
			case "public_key":
				weid.PublicKey, err = d.ReadString()
				if err != nil {
					return val, err
				}
			case "link_name":
				weid.LinkName, err = d.ReadString()
				if err != nil {
					return val, err
				}
			case "contract_id":
				weid.ContractId, err = d.ReadString()
				if err != nil {
					return val, err
				}
			default:
				d.Skip()
			}

			val.SetProvider(weid)
		}
	default:
		return val, errors.New("invalid wasmcloud entity type")
	}

	return val, nil
}

func MDecodeTraceContext(d *msgpack.Decoder) ([]wasmcloud_core.WasmcloudCoreTypesTraceContext, error) {
	isNil, err := d.IsNextNil()
	if err != nil || isNil {
		return []wasmcloud_core.WasmcloudCoreTypesTraceContext{}, err
	}
	size, err := d.ReadArraySize()
	if err != nil {
		return []wasmcloud_core.WasmcloudCoreTypesTraceContext{}, err
	}

	val := []wasmcloud_core.WasmcloudCoreTypesTraceContext{}
	for i := uint32(0); i < size; i++ {
		tT, err := MDecodeTuple2String(d)
		if err != nil {
			return []wasmcloud_core.WasmcloudCoreTypesTraceContext{}, err
		}
		val = append(val, tT)
	}
	return val, nil
}

func MDecodeTuple2String(d *msgpack.Decoder) (wasmcloud_core.WasmcloudCoreTypesTuple2StringStringT, error) {
	var val wasmcloud_core.WasmcloudCoreTypesTuple2StringStringT
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
		case "f0":
			val.F0, err = d.ReadString()
		case "f1":
			val.F1, err = d.ReadString()
		default:
			err = d.Skip()
		}
		if err != nil {
			return val, err
		}
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

func MDecodeLinkSettings(d *msgpack.Decoder) ([]wasmcloud_core.WasmcloudCoreTypesLinkSettings, error) {
	isNil, err := d.IsNextNil()
	if err != nil || isNil {
		return []wasmcloud_core.WasmcloudCoreTypesTuple2StringStringT{}, err
	}

	size, err := d.ReadArraySize()
	if err != nil {
		return []wasmcloud_core.WasmcloudCoreTypesTuple2StringStringT{}, err
	}

	val := []wasmcloud_core.WasmcloudCoreTypesLinkSettings{}
	for i := uint32(0); i < size; i++ {
		tT, err := MDecodeTuple2String(d)
		if err != nil {
			return []wasmcloud_core.WasmcloudCoreTypesTraceContext{}, err
		}
		val = append(val, tT)
	}
	return val, nil
}

func MDecodeHealthCheckResponse(d *msgpack.Decoder) (wasmcloud_core.WasmcloudCoreTypesHealthCheckResponse, error) {
	var val wasmcloud_core.WasmcloudCoreTypesHealthCheckResponse

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
		case "healthy":
			val.Healthy, err = d.ReadBool()
		case "message":
			val.Message, err = d.ReadString()
		default:
			err = d.Skip()
		}
		if err != nil {
			return val, err
		}
	}
	return val, nil
}
