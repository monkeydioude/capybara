package capybara

import (
	"fmt"
	"regexp"
	"strings"
)

// Method defines the way of matching a pattern defined in the
// config file against the URI requested. On a match, it returns
// the part of the URI following the pattern.
type Method func(string, string) (string, error)

// Methods is a map of Method with some comfy functions
type Methods map[string]Method

// Exists checks if method is set in map
func (ms Methods) Exists(name string) bool {
	_, ok := ms[name]
	return ok
}

// Add a method to the map
func (ms Methods) Add(name string, method Method) {
	ms[name] = method
}

func regex(pattern, URI string) (string, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	loc := r.FindStringIndex(URI)
	if loc == nil {
		return "", fmt.Errorf("could not regex match %s against %s", pattern, URI)
	}
	return URI[loc[1]:], nil
}

func str(pattern, URI string) (string, error) {
	if rest, ok := strings.CutPrefix(URI, pattern); ok {
		return rest, nil
	}
	return "", fmt.Errorf("could not string match %s against %s", pattern, URI)
}
