package main

import (
	"fmt"

	"github.com/wasmCloud/provider-sdk-go"
)

func main() {
	p, err := provider.New(
		provider.SourceLinkPut(handleNewSourceLink),
		provider.TargetLinkPut(handleNewTargetLink),
		provider.SourceLinkDel(handleDelSourceLink),
		provider.TargetLinkDel(handleDelTargetLink),
		provider.HealthCheck(handleHealthCheck),
		provider.Shutdown(handleShutdown),
	)
	if err != nil {
		panic(err)
	}

	err = p.Start()
	if err != nil {
		panic(err)
	}
}

func handleNewSourceLink(link provider.InterfaceLinkDefinition) error {
	fmt.Println("Received new source link", link)
	return nil
}

func handleNewTargetLink(link provider.InterfaceLinkDefinition) error {
	fmt.Println("Received new target link", link)
	return nil
}

func handleDelSourceLink(link provider.InterfaceLinkDefinition) error {
	fmt.Println("Received del source link", link)
	return nil
}

func handleDelTargetLink(link provider.InterfaceLinkDefinition) error {
	fmt.Println("Received del target link", link)
	return nil
}

func handleHealthCheck() string {
	fmt.Println("Received handle health check")
	return "provider healthy"
}

func handleShutdown() error {
	fmt.Println("Received handle shutdown")
	return nil
}
