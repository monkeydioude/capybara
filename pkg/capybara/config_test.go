package capybara

import (
	"testing"
)

func TestICanMakeConfigFromJsonFile(t *testing.T) {
	trial := NewConfig("../../testdata/dummy.config.json")

	s := []*service{
		&service{
			ID:      "test1",
			Pattern: "^/iron$",
			Port:    9090,
		},
		&service{
			ID:      "test2",
			Pattern: "/maiden",
			Method:  "string",
			Port:    9091,
		},
		&service{
			ID:      "test2",
			Pattern: "/maiden",
			Port:    9091,
		},
		&service{
			ID:       "test3",
			Pattern:  "/space?",
			Port:     9092,
			Method:   "string",
			Redirect: "/spaaaaaaaaaaaaaaaaaaaaaaace",
		},
	}

	goal := Config{
		Proxy: proxy{
			Port: 88,
		},
		Services: s,
	}

	assertEqual(t, *(trial.Services[0]), *(goal.Services[0]))
	assertEqual(t, *(trial.Services[1]), *(goal.Services[1]))
	assertNotEqual(t, *(trial.Services[1]), *(goal.Services[2]))
	assertEqual(t, *(trial.Services[2]), *(goal.Services[3]))
	assertEqual(t, trial.Proxy, trial.Proxy)
}
