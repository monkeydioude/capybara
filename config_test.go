package capybara

import (
	"testing"
)

func TestICanMakeConfigFromJsonFile(t *testing.T) {
	trial := NewConfig("testdata/dummy.config.json")

	s := []*service{&service{
		ID:      "pouet",
		Pattern: "^/pouet/",
		Port:    9090,
	},
	}

	goal := Config{
		Proxy: proxy{
			Port: 88,
		},
		Services: s,
	}

	assertEqual(t, *(trial.Services[0]), *(goal.Services[0]))
	assertEqual(t, trial.Proxy, trial.Proxy)
}
