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
