package capybara

import (
	"fmt"
	"net/http/httputil"
	"net/url"
)

type service struct {
	ID       string `json:"id"`
	Pattern  string `json:"pattern"`
	Method   string `json:"method"`
	Port     int    `json:"port"`
	Redirect string `json:"redirect"`
	Protocol string `json:"protocol"`
}

func (s *service) FixProtocol() {
	if s.Protocol == "" {
		s.Protocol = string(defaultProtocol)
	}
}

func (s *service) FixMethod() {
	// Unspecified method in json. Using default
	if s.Method == "" {
		s.Method = defaultMethod
	}
}

func (s *service) NewHttpReverseProxy(url *url.URL) (*httputil.ReverseProxy, error) {
	if url == nil {
		return nil, fmt.Errorf("%w: %w", ErrServiceHandleHTTP, ErrNilPointer)
	}
	return httputil.NewSingleHostReverseProxy(url), nil
}
