package capybara

import (
	"encoding/json"
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

func NewConfig(p string) *Config {
	d, err := os.ReadFile(p)

	if err != nil {
		log.Fatalf("[ERR ] Could not ReadFile, reason: %s", err)
	}
	c := &Config{}

	if err = json.Unmarshal(d, c); err != nil {
		if err = yaml.Unmarshal(d, c); err != nil {
			log.Fatalf("[ERR ] Could not Unmarshal config, reason: %s", err)
		}
	}

	return c
}
