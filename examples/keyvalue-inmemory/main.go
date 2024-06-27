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
	p := &Provider{
		sourceLinks:       make(map[string]provider.InterfaceLinkDefinition),
		targetLinks:       make(map[string]provider.InterfaceLinkDefinition),
		failedSourceLinks: make(map[string]provider.InterfaceLinkDefinition),
		failedTargetLinks: make(map[string]provider.InterfaceLinkDefinition),
	}

	wasmcloudprovider, err := provider.New(
		provider.SourceLinkPut(p.handleNewSourceLink),
		provider.TargetLinkPut(p.handleNewTargetLink),
		provider.SourceLinkDel(p.handleDelSourceLink),
		provider.TargetLinkDel(p.handleDelTargetLink),
		provider.HealthCheck(p.handleHealthCheck),
		provider.Shutdown(p.handleShutdown),
	)
	if err != nil {
		return err
	}

	providerCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	// Handle RPC operations
	stopFunc, err := server.Serve(wasmcloudprovider.RPCClient, p)
	if err != nil {
		wasmcloudprovider.Shutdown()
		return err
	}

	// Handle control interface operations
	go func() {
		err := wasmcloudprovider.Start()
		providerCh <- err
	}()

	// Shutdown on SIGINT
	signal.Notify(signalCh, syscall.SIGINT)

	select {
	case err = <-providerCh:
		stopFunc()
		return err
	case <-signalCh:
		wasmcloudprovider.Shutdown()
		stopFunc()
	}

	return nil
}
