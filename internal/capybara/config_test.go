package capybara

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestICanMakeConfigFromJsonFile(t *testing.T) {
	trial := NewConfig("../../testdata/dummy.config.json")

	s := []*service{
		{
			ID:      "test1",
			Pattern: "^/iron$",
			Port:    9090,
		},
		{
			ID:      "test2",
			Pattern: "/maiden",
			Method:  "string",
			Port:    9091,
		},
		{
			ID:      "test2",
			Pattern: "/maiden",
			Port:    9091,
		},
		{
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

	assert.Equal(t, *(trial.Services[0]), *(goal.Services[0]))
	assert.Equal(t, *(trial.Services[1]), *(goal.Services[1]))
	assert.NotEqual(t, *(trial.Services[1]), *(goal.Services[2]))
	assert.Equal(t, *(trial.Services[2]), *(goal.Services[3]))
	assert.Equal(t, trial.Proxy, trial.Proxy)
}
