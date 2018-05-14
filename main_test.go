package main

import (
	"testing"
)

func TestICanMakeConfigFromJsonFile(t *testing.T) {
	trial := newConfig("testdata/dummy.config.json")

	goal := config{
		Proxy: proxy{
			Port: 80,
		},
		Services: &[]service{
			service{
				ID:      "pouet",
				Pattern: "^/pouet/",
				Port:    9090,
			},
		},
	}

	assertEqual(t, (*trial.Services)[0], (*goal.Services)[0])
	assertEqual(t, trial.Proxy, trial.Proxy)
}
