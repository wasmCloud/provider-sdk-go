//go:generate wit-bindgen-wrpc go --out-dir bindings --package github.com/wasmCloud/provider-sdk-go/examples/http-server/bindings wit

package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.wasmcloud.dev/provider"
	"go.wasmcloud.dev/provider/wrpchttp"
)

type Server struct {
	wasiIncomingHandler *http.Client
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// NOTE(lxf): net/http doesn't allow 'RequestURI' to be present in outbound requests, so we clear it here.
	// https://go.dev/src/net/http/client.go
	r.RequestURI = ""

	resp, err := s.wasiIncomingHandler.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)

	for k, vals := range resp.Trailer {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
}

func main() {
	// NOTE(lxf): Enable wrpc debugging
	// lvl := new(slog.LevelVar)
	// lvl.Set(slog.LevelDebug)
	// logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	// 	Level: lvl,
	// }))
	// slog.SetDefault(logger)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func serveLocal(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from the provider!"))
}

func run() error {
	wasmcloudprovider, err := provider.New()
	if err != nil {
		return err
	}

	providerCh := make(chan error, 1)
	httpCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	proxyServer := &Server{
		wasiIncomingHandler: &http.Client{
			Transport: wrpchttp.NewIncomingRoundTripper(wasmcloudprovider, wrpchttp.WithSingleTarget("http-http_component")),
			Timeout:   time.Second * 5,
		},
	}

	// Handle control interface operations
	go func() {
		err := wasmcloudprovider.Start()
		providerCh <- err
	}()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/proxy", proxyServer)
		mux.Handle("/", http.HandlerFunc(serveLocal))
		err := http.ListenAndServe(":8080", mux)
		httpCh <- err
	}()

	// Shutdown on SIGINT
	signal.Notify(signalCh, syscall.SIGINT)

	select {
	case err = <-providerCh:
		return err
	case <-httpCh:
		wasmcloudprovider.Shutdown()
	case <-signalCh:
		wasmcloudprovider.Shutdown()
	}

	return nil
}
