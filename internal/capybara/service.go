package capybara

import (
	"fmt"

	ser "github.com/monkeydioude/capybara/internal/service"
)

type service struct {
	ID       string      `json:"id" yaml:"id"`
	Pattern  string      `json:"pattern" yaml:"pattern"`
	Method   string      `json:"method" yaml:"method"`
	Port     int32       `json:"port" yaml:"port"`
	Redirect string      `json:"redirect" yaml:"redirect"`
	Protocol Protocol    `json:"protocol" yaml:"protocol"`
	Host     string      `json:"host" yaml:"host"`
	Schema   *ser.Schema `json:"schema" yaml:"schema"`
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

func (s service) MatchHost(host string) bool {
	if s.Host == "" {
		return true
	}
	return host != "" && host == s.Host
}

func (s service) String() string {
	return fmt.Sprintf("\"%s\": %s => :%d", s.ID, s.Pattern, s.Port)
}
