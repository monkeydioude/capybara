package capybara

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

// LoadConfig reads and validates a JSON or YAML config file. It returns an error instead
// of exiting, so that a reload can keep the last good config.
func LoadConfig(p string) (*Config, error) {
	d, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	c := &Config{}
	if jerr := json.Unmarshal(d, c); jerr != nil {
		c = &Config{} // a failed JSON pass may have half-filled the struct
		if yerr := yaml.Unmarshal(d, c); yerr != nil {
			return nil, fmt.Errorf("parse config %s: not JSON (%v) nor YAML (%w)", p, jerr, yerr)
		}
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", p, err)
	}
	return c, nil
}

// NewConfig is LoadConfig for startup: a config that cannot be loaded is fatal.
func NewConfig(p string) *Config {
	c, err := LoadConfig(p)
	if err != nil {
		log.Fatalf("[ERR ] %s", err)
	}
	return c
}

// Validate refuses a config that would route nothing, crash the proxy, or silently expose
// a whole backend.
func (c *Config) Validate() error {
	if len(c.Services) == 0 {
		return errors.New("no services defined")
	}
	var errs []error
	for i, s := range c.Services {
		if s == nil {
			errs = append(errs, fmt.Errorf("services[%d]: empty entry", i))
			continue
		}
		errs = append(errs, s.validate(fmt.Sprintf("services[%d] %q", i, s.ID))...)
	}
	return errors.Join(errs...)
}
