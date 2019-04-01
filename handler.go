package capybara

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/monkeydioude/tools"
)

type proxy struct {
	Port int `json:"port"`
}

type service struct {
	ID            string `json:"id"`
	Pattern       string `json:"pattern"`
	Port          int    `json:"port"`
	RemovePattern bool   `json:"removePattern,omitempty"`
}

type Config struct {
	Proxy    proxy      `json:"proxy"`
	Services []*service `json:"services"`
}

type Handler struct {
	services []*service
}

func NewHandler(services []*service) *Handler {
	return &Handler{
		services: services,
	}
}

func buildURL(p int) string {
	var b strings.Builder
	b.WriteString("http://localhost:")
	b.WriteString(strconv.Itoa(p))

	return b.String()
}

// ServeHTTP implements net/http/Handler interface
func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/favicon.ico" {
		rw.WriteHeader(200)
		return
	}

	for _, service := range h.services {
		if _, err := tools.MatchAndFind(service.Pattern, r.RequestURI); err != nil {
			continue
		}

		u, err := url.Parse(buildURL(service.Port))
		if err != nil {
			go Log(fmt.Sprintf("[WARN] Could not parse url http://localhost:%d", service.Port))
			continue
		}

		rp := httputil.NewSingleHostReverseProxy(u)
		if service.RemovePattern {
			r.URL.Path = "/"
		}
		rp.ServeHTTP(rw, r)
		return
	}

	http.NotFound(rw, r)
}

func (s *service) String() string {
	return fmt.Sprintf("\"%s\": %s => :%d", s.ID, s.Pattern, s.Port)
}
