package main

import (
	"github.com/wasmCloud/provider-sdk-go"
)

func main() {
	p, err := provider.New(
	// provider.WithProviderActionFunc(handleKVAction),
	// provider.WithNewLinkFunc(handleNewLink),
	// provider.WithDelLinkFunc(handleDelLink),
	// provider.WithHealthCheckMsg(healthCheckMsg),
	)
	if err != nil {
		panic(err)
	}

	err = p.Start()
	if err != nil {
		panic(err)
	}
}
