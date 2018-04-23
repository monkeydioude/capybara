package main

import "testing"

func TestICanMatchAPatternToAStringAndRetrieveParts(t *testing.T) {
	pattern := "^/pouet/(.+)"
	trial := "/pouet/pwet"

	parts, err := matchAndFind(pattern, trial)

	if err != nil || len(parts) != 2 || parts[1] != "pwet" {
		t.Fail()
	}
}
