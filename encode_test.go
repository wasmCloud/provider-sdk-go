package provider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wasmCloud/provider-sdk-go"
	wasmcloud_core "github.com/wasmCloud/provider-sdk-go/core"
	msgpack "github.com/wasmcloud/tinygo-msgpack"
)

func TestEncodeDecodeHealthCheckResponse(t *testing.T) {
	a := wasmcloud_core.WasmcloudCoreTypesHealthCheckResponse{
		Healthy: true,
		Message: "health test",
	}

	buf := provider.MEncode(&a)

	assert.NotNil(t, buf)

	dec := msgpack.NewDecoder(buf)
	b, err := provider.MDecodeHealthCheckResponse(&dec)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}

func TestEncodeDecodeInvocation(t *testing.T) {
	// Invocation from actor to provider
	actor_entity := wasmcloud_core.WasmcloudCoreTypesWasmcloudEntity{}
	actor_entity.SetActor("iamanactor")

	provider_entity := wasmcloud_core.WasmcloudCoreTypesWasmcloudEntity{}
	provider_entity.SetProvider(wasmcloud_core.WasmcloudCoreTypesProviderIdentifier{
		PublicKey:  "iamsuperpublic",
		ContractId: "contract:topsecret",
		LinkName:   "default",
	})

	a := wasmcloud_core.WasmcloudCoreTypesInvocation{
		Origin:        actor_entity,
		Target:        provider_entity,
		Operation:     "Wasmcloud:Tester",
		Msg:           []byte("test"),
		Id:            "uuids-are-too-long",
		EncodedClaims: "beepboopimencoded",
		SourceHostId:  "SOMEHOSTID",
		ContentLength: 0,
		TraceContext: []wasmcloud_core.WasmcloudCoreTypesTuple2StringStringT{
			{F0: "test", F1: "derp"},
		},
	}

	buf := provider.MEncode(&a)

	assert.NotNil(t, buf)

	dec := msgpack.NewDecoder(buf)
	b, err := provider.MDecodeInvocation(&dec)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}

func TestEncodeDecodeWasmcloudEntityActor(t *testing.T) {
	a := wasmcloud_core.WasmcloudCoreTypesWasmcloudEntity{}
	a.SetActor("iamanactor")

	buf := provider.MEncode(&a)

	assert.NotNil(t, buf)

	dec := msgpack.NewDecoder(buf)
	b, err := provider.MDecodeWasmCloudEntity(&dec)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}

func TestEncodeDecodeWasmcloudEntityProvider(t *testing.T) {
	a := wasmcloud_core.WasmcloudCoreTypesWasmcloudEntity{}
	a.SetProvider(wasmcloud_core.WasmcloudCoreTypesProviderIdentifier{
		PublicKey:  "iamsuperpublic",
		ContractId: "contract:topsecret",
		LinkName:   "default",
	})

	buf := provider.MEncode(&a)

	assert.NotNil(t, buf)

	dec := msgpack.NewDecoder(buf)
	b, err := provider.MDecodeWasmCloudEntity(&dec)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}

func TestEncodeDecodeLinkDefinationsNone(t *testing.T) {
	val := wasmcloud_core.Option[[]wasmcloud_core.WasmcloudCoreTypesTuple2StringStringT]{}
	val.Unset()
	assert.True(t, val.IsNone())

	a := wasmcloud_core.WasmcloudCoreTypesLinkDefinition{
		ActorId:    "actor-id",
		ProviderId: "provider-id",
		LinkName:   "default",
		ContractId: "Wasmcloud:Tester",
		Values:     val,
	}

	buf := provider.MEncode(&a)

	assert.NotNil(t, buf)

	dec := msgpack.NewDecoder(buf)
	b, err := provider.MDecodeLinkDefinition(&dec)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}

func TestEncodeDecodeLinkDefinationsSome(t *testing.T) {
	val := wasmcloud_core.Option[[]wasmcloud_core.WasmcloudCoreTypesTuple2StringStringT]{}
	val.Set([]wasmcloud_core.WasmcloudCoreTypesTuple2StringStringT{{F0: "foo", F1: "bar"}})
	assert.True(t, val.IsSome())

	a := wasmcloud_core.WasmcloudCoreTypesLinkDefinition{
		ActorId:    "actor-id",
		ProviderId: "provider-id",
		LinkName:   "default",
		ContractId: "Wasmcloud:Tester",
		Values:     val,
	}

	buf := provider.MEncode(&a)
	assert.NotNil(t, buf)

	dec := msgpack.NewDecoder(buf)
	b, err := provider.MDecodeLinkDefinition(&dec)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}

func TestEncodeDecodeCapabilityContract(t *testing.T) {
	var a wasmcloud_core.WasmcloudCoreTypesCapabilityContractId = "Wasmcloud:Tester"

	buf := provider.MEncode(a)
	assert.NotNil(t, buf)

	dec := msgpack.NewDecoder(buf)
	b, err := provider.MDecodeCapabilityContractId(&dec)

	assert.NoError(t, err)
	assert.Equal(t, a, b)

}

func TestEncodeDecodeInvocationResponseNoError(t *testing.T) {
	iErr := wasmcloud_core.Option[string]{}
	iErr.Unset()
	a := wasmcloud_core.WasmcloudCoreTypesInvocationResponse{
		Msg:           []byte("stop reading my messages"),
		InvocationId:  "uuids-are-too-long",
		Error:         iErr,
		ContentLength: 0,
	}

	buf := provider.MEncode(&a)
	assert.NotNil(t, buf)

	dec := msgpack.NewDecoder(buf)
	b, err := provider.MDecodeInvocationResponse(&dec)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}

func TestEncodeDecodeInvocationResponseWithError(t *testing.T) {
	iErr := wasmcloud_core.Option[string]{}
	iErr.Set("critical failure")
	a := wasmcloud_core.WasmcloudCoreTypesInvocationResponse{
		Msg:           []byte("stop reading my messages"),
		InvocationId:  "uuids-are-too-long",
		Error:         iErr,
		ContentLength: 0,
	}

	buf := provider.MEncode(&a)
	assert.NotNil(t, buf)

	dec := msgpack.NewDecoder(buf)
	b, err := provider.MDecodeInvocationResponse(&dec)

	assert.NoError(t, err)
	assert.Equal(t, a, b)
}
