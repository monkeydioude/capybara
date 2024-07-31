package capybara

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

const defaultMethod = "regex"
const defaultLocalhost = "http://localhost"

type proxy struct {
	Port        int    `json:"port"`
	TLSCacheDir string `json:"tls_cache_dir,omitempty"`
	TLSHost     string `json:"tls_host,omitempty"`
}

type service struct {
	ID       string `json:"id"`
	Pattern  string `json:"pattern"`
	Method   string `json:"method"`
	Port     int    `json:"port"`
	Redirect string `json:"redirect"`
}

// Config handles the config fed to Capybara.
// Todo: config should check itself (<insert Ice Cube joke>) on startup
type Config struct {
	Proxy    proxy      `json:"proxy"`
	Services []*service `json:"services"`
}

// Handler take care of the matching pattern against route part of Capybara.
type Handler struct {
	services []*service
	Methods  Methods
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
		// Unspecified method in json. Using default
		if service.Method == "" {
			service.Method = defaultMethod
		}

		if !h.Methods.Exists(service.Method) {
			log.Printf("[WARN] Could not find method %s in methods' map", service.Method)
			continue
		}

		if err := h.Methods[service.Method](service.Pattern, r.RequestURI); err != nil {
			continue
		}

		u, err := url.Parse(buildURL(service.Port))
		if err != nil {
			go Log(fmt.Sprintf("[WARN] Could not parse url http://localhost:%d", service.Port))
			continue
		}

		rp := httputil.NewSingleHostReverseProxy(u)
		if service.Redirect != "" {
			r.URL.Path = service.Redirect
		}
		rp.ServeHTTP(rw, r)
		return
	}

	http.NotFound(rw, r)
}

func (s *service) String() string {
	return fmt.Sprintf("\"%s\": %s => :%d", s.ID, s.Pattern, s.Port)
}
