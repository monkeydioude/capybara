package capybara

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func NewConfig(p string) (c *Config) {
	d, err := ioutil.ReadFile(p)

	if err != nil {
		log.Fatalf("[ERR ] Could not ReadFile, reason: %s", err)
	}

	err = json.Unmarshal(d, &c)

	if err != nil {
		log.Fatalf("[ERR ] Could not Unmarshal config, reason: %s", err)
	}

	return
}
