package capybara

import (
	"fmt"

	"github.com/monkeydioude/tools"
)

// Method defines the way of matching a pattern defined in the
// config file against the URI requested
type Method func(string, string) error

// Methods is a map of Method with some comfy functions
type Methods map[string]Method

// Exists checks if method is set in map
func (ms Methods) Exists(name string) bool {
	if _, ok := ms[name]; !ok {
		return false
	}
	return true
}

// Add a method to the map
func (ms Methods) Add(name string, method Method) {
	ms[name] = method
}

func regex(pattern, URI string) error {
	_, err := tools.MatchAndFind(pattern, URI)

	return err
}

func str(pattern, URI string) error {
	if len(pattern) <= len(URI) && pattern == URI[:len(pattern)] {
		return nil
	}
	return fmt.Errorf("Could not string match %s against %s", pattern, URI)
}
