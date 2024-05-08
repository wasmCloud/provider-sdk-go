//go:generate wit-bindgen-wrpc go --out-dir bindings --package github.com/wasmCloud/provider-sdk-go/examples/keyvalue-inmemory/bindings wit

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wasmCloud/provider-sdk-go"
	server "github.com/wasmCloud/provider-sdk-go/examples/keyvalue-inmemory/bindings"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	p, err := provider.New(
		provider.SourceLinkPut(handleNewSourceLink),
		provider.TargetLinkPut(handleNewTargetLink),
		provider.SourceLinkDel(handleDelSourceLink),
		provider.TargetLinkDel(handleDelTargetLink),
		provider.HealthCheck(handleHealthCheck),
		provider.Shutdown(handleShutdown),
	)
	if err != nil {
		return err
	}
	providerCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	// Handle RPC operations
	stopFunc, err := server.Serve(p.RPCClient, &Provider{})
	if err != nil {
		p.Shutdown()
		return err
	}

	// Handle control interface operations
	go func() {
		err := p.Start()
		providerCh <- err
	}()

	// Shutdown on SIGINT
	signal.Notify(signalCh, syscall.SIGINT)

	select {
	case err = <-providerCh:
		stopFunc()
		return err
	case <-signalCh:
		p.Shutdown()
		stopFunc()
	}

	return nil
}

func handleNewSourceLink(link provider.InterfaceLinkDefinition) error {
	log.Println("Handling new source link", link)
	return nil
}

func handleNewTargetLink(link provider.InterfaceLinkDefinition) error {
	log.Println("Handling new target link", link)
	return nil
}

func handleDelSourceLink(link provider.InterfaceLinkDefinition) error {
	log.Println("Handling del source link", link)
	return nil
}

func handleDelTargetLink(link provider.InterfaceLinkDefinition) error {
	log.Println("Handling del target link", link)
	return nil
}

func handleHealthCheck() string {
	log.Println("Handling health check")
	return "provider healthy"
}

func handleShutdown() error {
	log.Println("Handling shutdown")
	return nil
}
