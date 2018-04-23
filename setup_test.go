package main

import "testing"

// This file is only here to setup tools for testing purpose

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}
