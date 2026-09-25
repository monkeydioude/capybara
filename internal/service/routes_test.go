package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoutesAllow(t *testing.T) {
	for _, tc := range []struct {
		name   string
		routes Routes
		method string
		route  string
		allow  bool
	}{
		{"no routes", Routes{}, "GET", "users", false},
		{"exact", Routes{"users": {"GET"}}, "GET", "users", true},
		{"exact, other method", Routes{"users": {"GET"}}, "POST", "users", false},
		{"method case", Routes{"users": {"get"}}, "GET", "users", true},
		{"any method", Routes{"users": {"*"}}, "OPTIONS", "users", true},
		{"leading slashes", Routes{"/users": {"GET"}}, "GET", "/users", true},
		{"exact ignores sub-routes", Routes{"users": {"GET"}}, "GET", "users/1", false},
		{"wildcard, self", Routes{"books*": {"GET"}}, "GET", "books", true},
		{"wildcard, sub-route", Routes{"books*": {"GET"}}, "GET", "books/1/pages", true},
		{"wildcard, same prefix", Routes{"books*": {"GET"}}, "GET", "booksellers", false},
		{"slash wildcard, self", Routes{"books/*": {"GET"}}, "GET", "books", false},
		{"slash wildcard, sub-route", Routes{"books/*": {"GET"}}, "GET", "books/1", true},
		{"catch-all", Routes{"*": {"GET"}}, "GET", "anything/at/all", true},
		{"catch-all, root", Routes{"*": {"GET"}}, "GET", "", true},
		{"dot dot", Routes{"books*": {"GET"}}, "GET", "books/../admin", false},
		{"dot", Routes{"books*": {"GET"}}, "GET", "books/./1", false},
		{"exact beats wildcard", Routes{"books": {"GET"}, "books*": {"*"}}, "POST", "books", false},
		{"wildcard still applies below", Routes{"books": {"GET"}, "books*": {"*"}}, "POST", "books/1", true},
		{"longer wildcard wins", Routes{"books*": {"*"}, "books/archive*": {"GET"}}, "POST", "books/archive/1", false},
		{"shorter wildcard elsewhere", Routes{"books*": {"*"}, "books/archive*": {"GET"}}, "POST", "books/1", true},
	} {
		assert.Equal(t, tc.allow, tc.routes.Allow(tc.method, tc.route), tc.name)
	}
}
