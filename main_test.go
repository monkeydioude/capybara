package main

import (
	"os"
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

func TestICanGetConfigPathFromArgs(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"", "./testdata/dummy.config.json"}

	cp, err := getConfigPath()

	if err != nil {
		t.Fatal(err)
	}

	assertEqual(t, cp, "./testdata/dummy.config.json")

	os.Args = oldArgs
}

func TestIFailIfNoConfigPathIsProvidedAsArgs(t *testing.T) {
	cp, err := getConfigPath()

	if cp != "" || err == nil {
		t.Fatalf("path should be empty (\"%s\" atm) and error shoud have been triggered (\"%v\" atm)\n", cp, err)
	}
}
