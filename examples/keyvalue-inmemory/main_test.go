package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/wasmCloud/provider-sdk-go/examples/keyvalue-inmemory/bindings/testing/wrpc/keyvalue/store"
	"go.wasmcloud.dev/provider"
	wrpcnats "wrpc.io/go/nats"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSet(t *testing.T) {
	env, err := NewTestEnvironment(context.Background(), t)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		hostDataSource, _ := env.HostDataSource()
		if err := run(hostDataSource); err != nil {
			log.Fatal(err)
		}
	}()
	// Give the provider a second to start
	env.EnsureProviderStarted()

	wrpc, err := env.WrpcClient()
	if err != nil {
		log.Fatal(err)
	}

	testBucket := "test-bucket"
	testKey := "test-key"
	testValue := "test-value"

	setCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	setResp, setErr := store.Set(setCtx, wrpc, testBucket, testKey, []byte(testValue))

	if setErr != nil {
		t.Errorf("`wrpc:keyvalue/store.set` failed unexpectedly: %v", err)
	}

	if setResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.set` returned error: %v", setResp.Err)
	}
}

func TestGet(t *testing.T) {
	env, err := NewTestEnvironment(context.Background(), t)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		hostDataSource, _ := env.HostDataSource()
		if err := run(hostDataSource); err != nil {
			log.Fatal(err)
		}
	}()
	// Give the provider a second to start
	env.EnsureProviderStarted()

	wrpc, err := env.WrpcClient()
	if err != nil {
		log.Fatal(err)
	}

	testBucket := "test-bucket"
	testKey := "test-key"
	testValue := "test-value"

	setCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	setResp, setErr := store.Set(setCtx, wrpc, testBucket, testKey, []byte(testValue))

	if setErr != nil {
		t.Errorf("`wrpc:keyvalue/store.set` failed unexpectedly: %v", setErr)
	}

	if setResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.set` returned error: %v", setResp.Err)
	}

	getCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	getResp, getErr := store.Get(getCtx, wrpc, testBucket, testKey)

	if getErr != nil {
		t.Errorf("`wrpc:keyvalue/store.get` failed unexpectedly: %v", getErr)
	}

	if getResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.get` returned error: %v", getResp.Err)
	}

	if string(*getResp.Ok) != string(testValue) {
		t.Errorf("want: %s, got: %s", string(testValue), string(*getResp.Ok))
	}
}

func TestExists(t *testing.T) {
	env, err := NewTestEnvironment(context.Background(), t)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		hostDataSource, _ := env.HostDataSource()
		if err := run(hostDataSource); err != nil {
			log.Fatal(err)
		}
	}()
	// Give the provider a second to start
	env.EnsureProviderStarted()

	wrpc, err := env.WrpcClient()
	if err != nil {
		log.Fatal(err)
	}

	testBucket := "test-bucket"
	testKey := "test-key"
	testValue := "test-value"

	setCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	setResp, setErr := store.Set(setCtx, wrpc, testBucket, testKey, []byte(testValue))

	if setErr != nil {
		t.Errorf("`wrpc:keyvalue/store.set` failed unexpectedly: %v", setErr)
	}

	if setResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.set` returned error: %v", setResp.Err)
	}

	existsCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	existsResp, existsErr := store.Exists(existsCtx, wrpc, testBucket, testKey)

	if existsErr != nil {
		t.Errorf("`wrpc:keyvalue/store.exists` failed unexpectedly: %v", existsErr)
	}

	if existsResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.exists` returned error: %v", existsResp.Err)
	}

	if *existsResp.Ok != true {
		t.Errorf("expected `wrpc:keyvalue/store.exists` to return true, but got %t", *existsResp.Ok)
	}

	doesntExistCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doesntExistResp, doesntExistErr := store.Exists(doesntExistCtx, wrpc, testBucket, fmt.Sprintf("%s-does-not-exist", testKey))

	if doesntExistErr != nil {
		t.Errorf("`wrpc:keyvalue/store.exists` failed unexpectedly: %v", doesntExistErr)
	}

	if doesntExistResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.exists` returned error: %v", doesntExistResp.Err)
	}

	if *doesntExistResp.Ok == true {
		t.Errorf("expected `wrpc:keyvalue/store.exists` to return false, but got %t", *doesntExistResp.Ok)
	}
}

func TestDelete(t *testing.T) {
	env, err := NewTestEnvironment(context.Background(), t)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		hostDataSource, _ := env.HostDataSource()
		if err := run(hostDataSource); err != nil {
			log.Fatal(err)
		}
	}()
	// Give the provider a second to start
	env.EnsureProviderStarted()

	wrpc, err := env.WrpcClient()
	if err != nil {
		log.Fatal(err)
	}

	testBucket := "test-bucket"
	testKey := "test-key"
	testValue := "test-value"

	setCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	setResp, setErr := store.Set(setCtx, wrpc, testBucket, testKey, []byte(testValue))

	if setErr != nil {
		t.Errorf("`wrpc:keyvalue/store.set` failed unexpectedly: %v", setErr)
	}

	if setResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.set` returned error: %v", setResp.Err)
	}

	existsCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	existsResp, existsErr := store.Exists(existsCtx, wrpc, testBucket, testKey)

	if existsErr != nil {
		t.Errorf("`wrpc:keyvalue/store.exists` failed unexpectedly: %v", existsErr)
	}

	if existsResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.exists` returned error: %v", existsResp.Err)
	}

	if *existsResp.Ok != true {
		t.Errorf("expected `wrpc:keyvalue/store.exists` to be true, but got %t", *existsResp.Ok)
	}

	deleteCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	deleteResp, deleteErr := store.Delete(deleteCtx, wrpc, testBucket, testKey)

	if deleteErr != nil {
		t.Errorf("`wrpc:keyvalue/store.delete` failed unexpectedly: %v", deleteErr)
	}

	if deleteResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.delete` returned error: %v", deleteResp.Err)
	}

	doesntExistCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	doesntExistResp, doesntExistErr := store.Exists(doesntExistCtx, wrpc, testBucket, testKey)

	if doesntExistErr != nil {
		t.Errorf("`wrpc:keyvalue/store.exists` failed unexpectedly: %v", doesntExistErr)
	}

	if doesntExistResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.exists` returned error: %v", doesntExistResp.Err)
	}

	if *doesntExistResp.Ok == true {
		t.Errorf("expected `wrpc:keyvalue/store.exists` to be false, but got %t", *doesntExistResp.Ok)
	}
}

func TestListKeys(t *testing.T) {
	env, err := NewTestEnvironment(context.Background(), t)
	// defer env.Cleanup()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		hostDataSource, _ := env.HostDataSource()
		if err := run(hostDataSource); err != nil {
			log.Fatal(err)
		}
	}()
	// Give the provider a second to start
	env.EnsureProviderStarted()

	wrpc, err := env.WrpcClient()
	if err != nil {
		log.Fatal(err)
	}

	testBucket := "test-bucket"
	testKey := "test-key"
	testValue := "test-value"
	testKeys := []string{}
	for i := range 5 {
		testKeys = append(testKeys, fmt.Sprintf("%s-%d", testKey, i))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for _, testKeyId := range testKeys {
		resp, setErr := store.Set(ctx, wrpc, "test-bucket", testKeyId, []byte(testValue))

		if setErr != nil {
			t.Errorf("`wrpc:keyvalue/store.set` failed unexpectedly: %v", err)
		}

		if resp.Err != nil {
			t.Errorf("`wrpc:keyvalue/store.set` returned error: %v", resp.Err)
		}
	}

	listKeysCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	listKeysResp, listKeysErr := store.ListKeys(listKeysCtx, wrpc, testBucket, nil)

	if listKeysErr != nil {
		t.Errorf("`wrpc:keyvalue/store.list-keys` failed unexpectedly: %v", listKeysErr)
	}

	if listKeysResp.Err != nil {
		t.Errorf("`wrpc:keyvalue/store.list-keys` returned error: %v", listKeysResp.Err)
	}

	for _, key := range testKeys {
		if !slices.Contains(listKeysResp.Ok.Keys, key) {
			t.Errorf("`wrpc:keyvalue/store.list-keys` response did not contain expected key: %s", key)
		}
	}
}

func NewTestEnvironment(ctx context.Context, t testing.TB) (*TestEnvironment, error) {
	natsServer, err := startContainer(ctx)
	normalizedName := strings.Replace(t.Name(), "/", "-", -1)
	testcontainers.CleanupContainer(t, natsServer.Container)
	if err != nil {
		return nil, err
	}
	hostData := provider.HostData{
		HostID:           "test-host",
		LatticeRPCPrefix: normalizedName,
		LatticeRPCURL:    natsServer.URI,
		ProviderKey:      normalizedName,
		// EnvValues:              map[string]string{},
		// InstanceID:             "",
		// ClusterIssuers:         []string{},
		// Config:                 map[string]string{},
		// Secrets:                map[string]provider.SecretValue{},
		// HostXKeyPublicKey: "",
		// ProviderXKeyPrivateKey: "",
		// HostXKeyPublicKey:      "",
		// LogLevel:               &"",
		// OtelConfig:             provider.OtelConfig{},
	}
	return &TestEnvironment{
		hostData: hostData,
		nats:     natsServer,
	}, nil
}

type TestEnvironment struct {
	hostData provider.HostData
	nats     *natsServer
}

func (te *TestEnvironment) EnsureProviderStarted() error {
	nc, ncErr := nats.Connect(te.hostData.LatticeRPCURL)
	if ncErr != nil {
		return fmt.Errorf("failed to connect to NATS: %v", ncErr)
	}
	for range 50 {
		time.Sleep(100 * time.Millisecond)
		_, err := nc.Request(te.ProviderHealthEndpoint(), nil, 100*time.Millisecond)
		if err == nil {
			break
		} else if err == nats.ErrNoResponders {
			continue
		} else {
			return err
		}
	}
	return nil
}

func (te *TestEnvironment) HostDataSource() (io.Reader, error) {
	hostDataJson, err := json.Marshal(te.hostData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HostData to json")
	}
	encoded := strings.NewReader(base64.StdEncoding.EncodeToString(hostDataJson))
	return encoded, nil
}

func (te *TestEnvironment) ProviderHealthEndpoint() string {
	return fmt.Sprintf("wasmbus.rpc.%s.health", te.ProviderNamespace())
}

func (te *TestEnvironment) ProviderNamespace() string {
	return fmt.Sprintf("%s.%s", te.hostData.LatticeRPCPrefix, te.hostData.ProviderKey)
}

func (te *TestEnvironment) WrpcClient() (*wrpcnats.Client, error) {
	nc, err := nats.Connect(te.hostData.LatticeRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %v", err)
	}
	return wrpcnats.NewClient(nc, wrpcnats.WithPrefix(te.ProviderNamespace())), nil
}

type natsServer struct {
	testcontainers.Container
	URI string
}

func startContainer(ctx context.Context) (*natsServer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "library/nats",
		ExposedPorts: []string{"4222/tcp"},
		WaitingFor:   wait.ForLog("Server is ready").WithStartupTimeout(10 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	var nats *natsServer
	if container != nil {
		nats = &natsServer{Container: container}
	}
	if err != nil {
		return nats, err
	}

	natsIp, err := container.Host(ctx)
	if err != nil {
		return nats, err
	}

	natsPort, err := container.MappedPort(ctx, "4222")
	if err != nil {
		return nats, err
	}

	nats.URI = fmt.Sprintf("nats://%s:%s", natsIp, natsPort.Port())
	return nats, nil
}
