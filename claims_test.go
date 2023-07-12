package provider_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wasmCloud/provider-sdk-go"
	wasmcloud_core "github.com/wasmCloud/provider-sdk-go/core"
)

func TestEncodeClaims(t *testing.T) {
	guid := uuid.NewString()

	prov := wasmcloud_core.WasmcloudCoreWasmcloudEntity{}
	prov.SetProvider(wasmcloud_core.WasmcloudCoreTypesProviderIdentifier{
		PublicKey:  "i-am-provider-key",
		LinkName:   "default",
		ContractId: "wasmcloud:tester",
	})

	actor := wasmcloud_core.WasmcloudCoreWasmcloudEntity{}
	actor.SetActor("i-am-actor-key")

	i := wasmcloud_core.WasmcloudCoreTypesInvocation{
		Origin:        prov,
		Target:        actor,
		Operation:     "Derper.Derp",
		Msg:           []byte("yolo"),
		Id:            guid,
		SourceHostId:  "i-am-host-id",
		ContentLength: uint64(len("yolo")),
	}

	hd := wasmcloud_core.WasmcloudCoreTypesHostData{
		HostId:         "i-am-host-id",
		InvocationSeed: "SCAJM3XPPWWGONF6JRV6ELRZQCXTEDZRX7RMVEWACCGXXYQP6N3SXI2IOM",
		LinkName:       "default",
	}

	err := provider.EncodeClaims(&i, hd, guid)
	assert.NoError(t, err)

	var claims provider.Claims

	_, err = jwt.ParseWithClaims(i.EncodedClaims, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("SCAJM3XPPWWGONF6JRV6ELRZQCXTEDZRX7RMVEWACCGXXYQP6N3SXI2IOM"), nil
	})

	// assert.NoError(t, err)

	b, err := json.MarshalIndent(claims, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(b))

	assert.Equal(t, "68B0A4763B1073A1B961B3EF8E8128BC38A65906310C074606CFA2193FE848D6", claims.Wascap.Hash)
}
