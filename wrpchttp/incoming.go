package wrpchttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	wasitypes "go.wasmcloud.dev/provider/internal/wasi/http/types"
	"go.wasmcloud.dev/provider/internal/wrpc/http/incoming_handler"
	wrpctypes "go.wasmcloud.dev/provider/internal/wrpc/http/types"

	wrpc "wrpc.io/go"
)

type IncomingRoundTripper struct {
	director    func(*http.Request) string
	natsCreator NatsClientCreator
	invoker     func(context.Context, wrpc.Invoker, *wrpctypes.Request) (*wrpc.Result[incoming_handler.Response, incoming_handler.ErrorCode], <-chan error, error)
}

var _ http.RoundTripper = (*IncomingRoundTripper)(nil)

type IncomingHandlerOption func(*IncomingRoundTripper)

func WithDirector(director func(*http.Request) string) IncomingHandlerOption {
	return func(p *IncomingRoundTripper) {
		p.director = director
	}
}

func WithSingleTarget(target string) IncomingHandlerOption {
	return WithDirector(func(_ *http.Request) string {
		return target
	})
}

func NewIncomingRoundTripper(nc NatsClientCreator, opts ...IncomingHandlerOption) *IncomingRoundTripper {
	p := &IncomingRoundTripper{
		natsCreator: nc,
		invoker:     incoming_handler.Handle,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *IncomingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	target := p.director(r)
	if target == "" {
		return nil, ErrNoTarget
	}

	outgoingBodyTrailer := HttpBodyToWrpc(r.Body, r.Trailer)
	pathWithQuery := r.URL.Path
	if r.URL.RawQuery != "" {
		pathWithQuery += "?" + r.URL.RawQuery
	}
	wreq := &wrpctypes.Request{
		Headers:       HttpHeaderToWrpc(r.Header),
		Method:        HttpMethodToWrpc(r.Method),
		Scheme:        HttpSchemeToWrpc(r.URL.Scheme),
		PathWithQuery: &pathWithQuery,
		Authority:     &r.Host,
		Body:          outgoingBodyTrailer,
		Trailers:      outgoingBodyTrailer,
	}

	wrpcClient := p.natsCreator.OutgoingRpcClient(target)
	wresp, errCh, err := p.invoker(r.Context(), wrpcClient, wreq)
	if err != nil {
		return nil, err
	}

	if wresp.Err != nil {
		return nil, fmt.Errorf("%w: %s", ErrRPC, wresp.Err)
	}

	respBody, trailers := WrpcBodyToHttp(wresp.Ok.Body, wresp.Ok.Trailers)

	resp := &http.Response{
		StatusCode: int(wresp.Ok.Status),
		Header:     make(http.Header),
		Request:    r,
		Body:       respBody,
		Trailer:    trailers,
	}

	for _, hdr := range wresp.Ok.Headers {
		for _, hdrVal := range hdr.V1 {
			resp.Header.Add(hdr.V0, string(hdrVal))
		}
	}

	errList := []error{}
	for err := range errCh {
		errList = append(errList, err)
	}

	if len(errList) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrRPC, errList)
	}

	return resp, nil
}

type wrpcIncomingBody struct {
	body           io.Reader
	trailer        http.Header
	trailerRx      wrpc.Receiver[[]*wrpc.Tuple2[string, [][]byte]]
	trailerOnce    sync.Once
	trailerIsReady uint32
}

func (r *wrpcIncomingBody) Close() error {
	return nil
}

func (r *wrpcIncomingBody) readTrailerOnce() {
	r.trailerOnce.Do(func() {
		trailers, err := r.trailerRx.Receive()
		if err != nil {
			return
		}
		for _, header := range trailers {
			for _, value := range header.V1 {
				r.trailer.Add(header.V0, string(value))
			}
		}
		atomic.CompareAndSwapUint32(&r.trailerIsReady, 0, 1)
	})
}

func (r *wrpcIncomingBody) Read(b []byte) (int, error) {
	n, err := r.body.Read(b)
	if err == io.EOF {
		r.readTrailerOnce()
	}
	return n, err
}

type wrpcOutgoingBody struct {
	body        io.ReadCloser
	trailer     http.Header
	bodyIsDone  chan struct{}
	trailerOnce sync.Once
}

func (r *wrpcOutgoingBody) Read(b []byte) (int, error) {
	n, err := r.body.Read(b)
	if err == io.EOF {
		r.finish()
	}
	return n, err
}

func (r *wrpcOutgoingBody) Receive() ([]*wrpc.Tuple2[string, [][]byte], error) {
	<-r.bodyIsDone
	trailers := HttpHeaderToWrpc(r.trailer)
	return trailers, nil
}

func (r *wrpcOutgoingBody) finish() {
	r.trailerOnce.Do(func() {
		r.body.Close()
		close(r.bodyIsDone)
	})
}

func (r *wrpcOutgoingBody) Close() error {
	r.finish()

	return nil
}

func HttpBodyToWrpc(body io.ReadCloser, trailer http.Header) *wrpcOutgoingBody {
	return &wrpcOutgoingBody{
		body:       body,
		trailer:    trailer,
		bodyIsDone: make(chan struct{}, 1),
	}
}

func WrpcBodyToHttp(body io.Reader, trailerRx wrpc.Receiver[[]*wrpc.Tuple2[string, [][]uint8]]) (*wrpcIncomingBody, http.Header) {
	trailer := make(http.Header)
	return &wrpcIncomingBody{
		body:      body,
		trailerRx: trailerRx,
		trailer:   trailer,
	}, trailer
}

func HttpMethodToWrpc(method string) *wrpctypes.Method {
	switch method {
	case http.MethodConnect:
		return wasitypes.NewMethodConnect()
	case http.MethodGet:
		return wasitypes.NewMethodGet()
	case http.MethodHead:
		return wasitypes.NewMethodHead()
	case http.MethodPost:
		return wasitypes.NewMethodPost()
	case http.MethodPut:
		return wasitypes.NewMethodPut()
	case http.MethodPatch:
		return wasitypes.NewMethodPatch()
	case http.MethodDelete:
		return wasitypes.NewMethodDelete()
	case http.MethodOptions:
		return wasitypes.NewMethodOptions()
	case http.MethodTrace:
		return wasitypes.NewMethodTrace()
	default:
		return wasitypes.NewMethodOther(method)
	}
}

func HttpSchemeToWrpc(scheme string) *wrpctypes.Scheme {
	switch scheme {
	case "http":
		return wasitypes.NewSchemeHttp()
	case "https":
		return wasitypes.NewSchemeHttps()
	default:
		return wasitypes.NewSchemeOther(scheme)
	}
}

func HttpHeaderToWrpc(header http.Header) []*wrpc.Tuple2[string, [][]uint8] {
	wasiHeader := make([]*wrpc.Tuple2[string, [][]uint8], 0, len(header))
	for k, vals := range header {
		var uintVals [][]uint8
		for _, v := range vals {
			uintVals = append(uintVals, []byte(v))
		}
		wasiHeader = append(wasiHeader, &wrpc.Tuple2[string, [][]uint8]{
			V0: k,
			V1: uintVals,
		})
	}

	return wasiHeader
}
