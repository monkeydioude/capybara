package capybara

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/monkeydioude/capybara/pkg/capybara/grpc"
)

const defaultMethod = "regex"
const defaultLocalhost = "http://localhost"

type proxy struct {
	Port        int    `json:"port"`
	TLSCacheDir string `json:"tls_cache_dir,omitempty"`
	TLSHost     string `json:"tls_host,omitempty"`
}

// Config handles the config fed to Capybara.
// Todo: config should check itself (<insert Ice Cube joke>) on startup
type Config struct {
	Proxy    proxy      `json:"proxy"`
	Services []*service `json:"services"`
}

// Handler take care of the matching pattern against route part of Capybara.
type Handler struct {
	services    []*service
	Methods     Methods
	certificate *tls.Certificate
}

func (h *Handler) SetCertificate(cert *tls.Certificate) {
	h.certificate = cert
}

// NewHandler gets feed a map of *service and "procude" a *Handler.
// That's how capybaras work.
func NewHandler(services []*service) *Handler {
	ms := make(Methods)
	ms.Add("string", str)
	ms.Add("regex", regex)

	return &Handler{
		services: services,
		Methods:  ms,
	}
}

func buildURL(p int) string {
	b := &strings.Builder{}

	b.WriteString(defaultLocalhost)
	b.WriteString(":")
	b.WriteString(strconv.Itoa(p))

	return b.String()
}

func (h *Handler) handleProtocol(rw http.ResponseWriter, r *http.Request, service *service, u *url.URL) error {
	if service == nil || u == nil {
		return ErrNilPointer
	}
	if grpc.IsGRPCRequest(r) {
		grpcServer, err := grpc.NewGRPCServer(h.certificate)
		if err != nil {
			return err
		}
		grpcServer.ServeHTTP(rw, r)
	} else {
		rp, err := service.NewHttpReverseProxy(u)
		if err != nil {
			return err
		}
		if service.Redirect != "" {
			r.URL.Path = service.Redirect
		}
		rp.ServeHTTP(rw, r)
	}
	return nil
}

// ServeHTTP implements net/http/Handler interface
func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/favicon.ico" {
		rw.WriteHeader(200)
		return
	}

	if r.RequestURI == "/healthcheck" {
		HealthcheckHandler(rw, r)
		return
	}

	for _, service := range h.services {
		service.FixMethod()
		service.FixProtocol()

		if !h.Methods.Exists(service.Method) {
			log.Printf("[WARN] Could not find method %s in methods' map", service.Method)
			continue
		}

		if err := h.Methods[service.Method](service.Pattern, r.RequestURI); err != nil {
			log.Printf("[WARN] Could not serve method %s with pattern %s", service.Method, service.Pattern)
			continue
		}

		u, err := url.Parse(buildURL(service.Port))
		if err != nil {
			go Log(fmt.Sprintf("[WARN] Could not parse url http://localhost:%d", service.Port))
			continue
		}

		if err := h.handleProtocol(rw, r, service, u); err != nil {
			go Log(fmt.Sprintf("[ERR ] Could not handle request http://localhost:%d, with url %s", service.Port, u))
			continue
		}
		return
	}

	http.NotFound(rw, r)
}

func (s *service) String() string {
	return fmt.Sprintf("\"%s\": %s => :%d", s.ID, s.Pattern, s.Port)
}
