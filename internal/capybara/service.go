package capybara

import (
	"fmt"
	"regexp"

	ser "github.com/monkeydioude/capybara/internal/service"
)

type service struct {
	ID       string     `json:"id" yaml:"id"`
	Pattern  string     `json:"pattern" yaml:"pattern"`
	Method   string     `json:"method" yaml:"method"`
	Port     int32      `json:"port" yaml:"port"`
	Redirect string     `json:"redirect" yaml:"redirect"`
	Protocol Protocol   `json:"protocol" yaml:"protocol"`
	Host     string     `json:"host" yaml:"host"`
	Routes   ser.Routes `json:"routes" yaml:"routes"`
}

func (s *service) FixProtocol() {
	if s.Protocol == "" {
		s.Protocol = defaultProtocol
	}
}

func (s *service) FixMethod() {
	// Unspecified method in json. Using default
	if s.Method == "" {
		s.Method = defaultMethod
	}
}

// validate refuses a service that would match nothing, match everything, or crash the proxy.
func (s *service) validate(name string) []error {
	var errs []error
	// An empty pattern matches every request: fine for a per-host catch-all, a config
	// mistake otherwise (use pattern "/" for a deliberate catch-all).
	if s.Pattern == "" && s.Host == "" {
		errs = append(errs, fmt.Errorf("%s: pattern is required unless host is set", name))
	}
	if s.Port < 1 || s.Port > 65535 {
		errs = append(errs, fmt.Errorf("%s: port %d is not in 1-65535", name, s.Port))
	}
	method := s.Method
	if method == "" {
		method = defaultMethod
	}
	switch method {
	case "string":
	case "regex":
		if _, err := regexp.Compile(s.Pattern); err != nil {
			errs = append(errs, fmt.Errorf("%s: pattern is not a valid regex: %w", name, err))
		}
	default:
		errs = append(errs, fmt.Errorf("%s: unknown method %q (string or regex)", name, s.Method))
	}
	switch s.Protocol {
	case "", HttpProtocol, RpcProtocol, WebSocketProtocol:
	default:
		errs = append(errs, fmt.Errorf("%s: unknown protocol %q (http, rpc or ws)", name, s.Protocol))
	}
	return errs
}

func (s service) MatchHost(host string) bool {
	if s.Host == "" {
		return true
	}
	return host != "" && host == s.Host
}

func (s service) String() string {
	return fmt.Sprintf("\"%s\": %s => :%d", s.ID, s.Pattern, s.Port)
}
