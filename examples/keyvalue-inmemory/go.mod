module github.com/wasmCloud/provider-sdk-go/examples/keyvalue-inmemory

go 1.22.3

require (
	github.com/wasmCloud/provider-sdk-go v0.0.0-20240124183610-1a92f8d04935
	github.com/wrpc/wrpc/go v0.0.0-20240723002736-26b614676513
)

require (
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/nats-io/nats.go v1.36.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
)

replace github.com/wasmCloud/provider-sdk-go => ../..
