package capybara

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/monkeydioude/capybara/internal/errors"
	capyGrpc "github.com/monkeydioude/capybara/internal/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const defaultProtocol = HttpProtocol

// backendTransport is shared by every HTTP and WebSocket service. It bounds what a backend
// that never answers can hold (dial and response-header timeouts) and keeps a pool of idle
// connections per backend. No cap on connections per backend: a cap makes the extra requests
// wait silently, which is worse than letting the client connection cap do its job.
var backendTransport = &http.Transport{
	DialContext:           (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
	MaxIdleConns:          256,
	MaxIdleConnsPerHost:   64,
	IdleConnTimeout:       90 * time.Second,
	ResponseHeaderTimeout: 5 * time.Minute, // long-polls included; a client that leaves cancels earlier
	ExpectContinueTimeout: time.Second,
}

type Protocol string

const (
	HttpProtocol      Protocol = "http"
	RpcProtocol       Protocol = "rpc"
	WebSocketProtocol Protocol = "ws"

	// Http3Protocol = "http3"
)

func (p Protocol) MatchesRaw(rawProto string) bool {
	return string(p) == rawProto
}

func (p Protocol) Matches(proto Protocol) bool {
	return p == proto
}

func (p Protocol) String() string {
	return string(p)
}

func FindOutProtocol(r *http.Request) Protocol {
	if capyGrpc.IsGRPCRequest(r) {
		return RpcProtocol
	}
	return defaultProtocol
}

func NewWebSocketProxy(url *url.URL) (*httputil.ReverseProxy, error) {
	if url == nil {
		return nil, fmt.Errorf("%w: %w", errors.ErrServiceHandleHTTP, errors.ErrNilPointer)
	}
	rp := httputil.NewSingleHostReverseProxy(url)
	rp.Transport = backendTransport
	rp.Director = func(req *http.Request) {
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
	}
	return rp, nil
}

func NewHttpReverseProxy(url *url.URL) (*httputil.ReverseProxy, error) {
	if url == nil {
		return nil, fmt.Errorf("%w: %w", errors.ErrServiceHandleHTTP, errors.ErrNilPointer)
	}
	rp := httputil.NewSingleHostReverseProxy(url)
	rp.Transport = backendTransport
	rp.Director = func(req *http.Request) {
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Origin-Host", url.Host)
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
	}
	return rp, nil
}

func NewGRPCServer(creds credentials.TransportCredentials, service *service) (*grpc.Server, error) {
	if creds == nil || service == nil {
		return nil, fmt.Errorf("NewGRPCServer: %w", errors.ErrNilPointer)
	}
	return grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnknownServiceHandler((&capyGrpc.CatchAllProxy{
			BackendPort: service.Port,
			Credentials: creds,
		}).Proxy),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    5 * time.Second, // Ping every 10 seconds
			Timeout: 5 * time.Second, // Wait 5 seconds for a ping response
		}),
	), nil
}
