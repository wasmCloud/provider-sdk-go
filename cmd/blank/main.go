package main

import (
	"fmt"

	"github.com/wasmCloud/provider-sdk-go"
)

func main() {
	fmt.Println("Initializing provider ...")
	p, err := provider.New(
		// provider.WithProviderActionFunc(handleKVAction),
		// provider.WithNewLinkFunc(handleNewLink),
		// provider.WithDelLinkFunc(handleDelLink),
		// provider.WithHealthCheckMsg(healthCheckMsg),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting provider ...")
	err = p.Start()
	if err != nil {
		panic(err)
	}
}
