package wrpchttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	wasitypes "go.wasmcloud.dev/provider/internal/wasi/http/types"
	"go.wasmcloud.dev/provider/internal/wrpc/http/incoming_handler"
	wrpctypes "go.wasmcloud.dev/provider/internal/wrpc/http/types"
	wrpc "wrpc.io/go"
	wrpcnats "wrpc.io/go/nats"
)

func TestHttpSchemeToWrpc(t *testing.T) {
	tt := map[string]wasitypes.SchemeDiscriminant{
		"http":              wasitypes.SchemeHttp,
		"https":             wasitypes.SchemeHttps,
		"some_other_scheme": wasitypes.SchemeOther,
	}

	for stdMethod, wasiMethod := range tt {
		t.Run(stdMethod, func(t *testing.T) {
			if want, got := wasiMethod, HttpSchemeToWrpc(stdMethod).Discriminant(); got != want {
				t.Errorf("want %v, got %v", want, got)
			}
		})
	}
}

func TestHttpMethodToWrpc(t *testing.T) {
	tt := map[string]wasitypes.MethodDiscriminant{
		http.MethodGet:     wasitypes.MethodGet,
		http.MethodHead:    wasitypes.MethodHead,
		http.MethodPost:    wasitypes.MethodPost,
		http.MethodPut:     wasitypes.MethodPut,
		http.MethodPatch:   wasitypes.MethodPatch,
		http.MethodDelete:  wasitypes.MethodDelete,
		http.MethodConnect: wasitypes.MethodConnect,
		http.MethodOptions: wasitypes.MethodOptions,
		http.MethodTrace:   wasitypes.MethodTrace,
	}

	for stdMethod, wasiMethod := range tt {
		t.Run(stdMethod, func(t *testing.T) {
			if want, got := wasiMethod, HttpMethodToWrpc(stdMethod).Discriminant(); got != want {
				t.Errorf("want %v, got %v", want, got)
			}
		})
	}
}

func TestHttpHeaderToWrpc(t *testing.T) {
	tt := map[string]http.Header{
		"blank": {},
		"single": {
			"Key": []string{"value"},
		},
		"multi": {
			"Key":     []string{"value1", "value2"},
			"Another": []string{"value1", "value2"},
		},
	}

	for name, headers := range tt {
		t.Run(name, func(t *testing.T) {
			wheaders := HttpHeaderToWrpc(headers)
			if want := len(headers); len(wheaders) != want {
				t.Errorf("want %v, got %v", want, len(wheaders))
			}

			for _, wheader := range wheaders {
				origHeader := headers.Values(wheader.V0)
				if want, got := len(origHeader), len(wheader.V1); got != want {
					t.Errorf("header '%s' expected %v values, got %v", wheader.V0, want, got)
					continue
				}
			}
		})
	}
}

type fakeNatsCreator struct {
	OutgoingRpcClientFunc func(target string) *wrpcnats.Client
}

func (f fakeNatsCreator) OutgoingRpcClient(target string) *wrpcnats.Client {
	return f.OutgoingRpcClientFunc(target)
}

type fakeReceiver struct {
	headers http.Header
}

func (f fakeReceiver) Receive() ([]*wrpc.Tuple2[string, [][]uint8], error) {
	return HttpHeaderToWrpc(f.headers), nil
}

func (fakeReceiver) Close() error {
	return nil
}

func TestRoundtrip(t *testing.T) {
	reqBody := "hello request"
	respBody := "hello response"
	pathWithQuery := "/path?q=val"
	req, _ := http.NewRequest(http.MethodPost, "http://example.com"+pathWithQuery, bytes.NewReader([]byte(reqBody)))
	req.Header.Add("X-Client-Custom", "x-client-value")

	wrpcTarget := "component_id"
	fakeNc := fakeNatsCreator{
		OutgoingRpcClientFunc: func(target string) *wrpcnats.Client {
			if target != wrpcTarget {
				t.Errorf("expected target %s, got %s", wrpcTarget, target)
			}
			return nil
		},
	}

	fakeInvoker := func(_ context.Context, _ wrpc.Invoker, wrpcReq *wrpctypes.Request) (*wrpc.Result[incoming_handler.Response, incoming_handler.ErrorCode], <-chan error, error) {
		if want, got := "example.com", *wrpcReq.Authority; want != got {
			t.Errorf("expected authority %s, got %s", want, got)
		}

		if want, got := pathWithQuery, *wrpcReq.PathWithQuery; want != got {
			t.Errorf("expected pathWithQuery %s, got %s", want, got)
		}

		if want, got := wasitypes.SchemeHttp, wrpcReq.Scheme.Discriminant(); want != got {
			t.Errorf("expected scheme %d, got %d", want, got)
		}

		if want, got := wasitypes.MethodPost, wrpcReq.Method.Discriminant(); want != got {
			t.Errorf("expected method %d, got %d", want, got)
		}

		if want, got := 1, len(wrpcReq.Headers); want != got {
			t.Fatalf("expected %v headers, got %v", want, got)
		}

		if want, got := "X-Client-Custom", wrpcReq.Headers[0].V0; want != got {
			t.Fatalf("expected header %v, got %v", want, got)
		}

		if want, got := 1, len(wrpcReq.Headers[0].V1); want != got {
			t.Fatalf("expected %v header values, got %v", want, got)
		}

		if want, got := "x-client-value", string(wrpcReq.Headers[0].V1[0]); want != got {
			t.Fatalf("expected header value %v, got %v", want, got)
		}

		body, err := io.ReadAll(wrpcReq.Body)
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		if want, got := reqBody, string(body); want != got {
			t.Errorf("expected body %v, got %v", want, got)
		}

		resp := wrpctypes.Response{
			Status:   http.StatusOK,
			Headers:  HttpHeaderToWrpc(http.Header{"X-Custom": []string{"x-value"}}),
			Body:     io.NopCloser(bytes.NewReader([]byte(respBody))),
			Trailers: fakeReceiver{headers: http.Header{}},
		}

		errCh := make(chan error)
		close(errCh)
		return wrpc.Ok[incoming_handler.ErrorCode](resp), errCh, nil
	}

	roundTripper := NewIncomingRoundTripper(fakeNc, WithSingleTarget(wrpcTarget))
	// NOTE(lxf): We are testing the roundtripper and not wrpc e2e
	roundTripper.invoker = fakeInvoker

	resp, err := roundTripper.RoundTrip(req)
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if want, got := http.StatusOK, resp.StatusCode; want != got {
		t.Errorf("expected status code %v, got %v", want, got)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if want, got := respBody, string(body); want != got {
		t.Errorf("expected body %v, got %v", want, got)
	}

	if want, got := error(nil), resp.Body.Close(); want != got {
		t.Errorf("expected body.Close() to return %v, got %v", want, got)
	}
}
