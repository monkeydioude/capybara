package service

import (
	"slices"
	"strings"
)

// wildcard matches any method when used in a method list, and any
// sub-route when used at the end of a route.
const wildcard = "*"

// Routes lists the public routes of a service, each with its allowed methods.
// Routes are relative to the service pattern:
//   - "users" only matches "users"
//   - "books*" matches "books" and everything under it ("books/1/pages"),
//     but not "booksellers"
//   - "*" matches every route
//
// When several routes match, the most specific one decides: an exact route
// beats a wildcard one, and a longer wildcard beats a shorter one.
type Routes map[string][]string

// Allow tells if method is allowed on route.
func (rs Routes) Allow(method, route string) bool {
	route = strings.TrimPrefix(route, "/")
	// "books/../admin" would pass as "books*" but reach "admin" once resolved
	if hasDotSegment(route) {
		return false
	}
	methods, ok := rs.lookup(route)
	return ok && slices.ContainsFunc(methods, func(m string) bool {
		return m == wildcard || strings.EqualFold(m, method)
	})
}

func (rs Routes) lookup(route string) ([]string, bool) {
	var best []string
	bestLen := -1

	for pattern, methods := range rs {
		pattern = strings.TrimPrefix(pattern, "/")
		prefix, isWildcard := strings.CutSuffix(pattern, wildcard)
		if !isWildcard {
			if pattern == route {
				return methods, true
			}
			continue
		}
		if len(prefix) > bestLen && isUnder(route, prefix) {
			best, bestLen = methods, len(prefix)
		}
	}
	return best, bestLen >= 0
}

// isUnder tells if route is prefix itself or one of its sub-routes.
func isUnder(route, prefix string) bool {
	rest, ok := strings.CutPrefix(route, prefix)
	if !ok {
		return false
	}
	return rest == "" || prefix == "" || strings.HasSuffix(prefix, "/") || rest[0] == '/'
}

func hasDotSegment(route string) bool {
	return slices.ContainsFunc(strings.Split(route, "/"), func(s string) bool {
		return s == "." || s == ".."
	})
}
