package main

import (
	"fmt"
	"regexp"
)

func matchAndFind(pattern, target string) ([]string, error) {
	r, err := regexp.Compile(pattern)

	if err != nil {
		return nil, fmt.Errorf("[WARN] %s", err)
	}

	if !r.MatchString(target) {
		return nil, fmt.Errorf("[WARN] Target '%s' did not match against '%s'", target, pattern)
	}

	return r.FindStringSubmatch(target), nil
}
