package service

import (
	"net/http"
	"slices"
)

type Route = string
type Method string

const (
	MethodGet    Method = http.MethodGet
	MethodPost   Method = http.MethodPost
	MethodPut    Method = http.MethodPut
	MethodPatch  Method = http.MethodPatch
	MethodDelete Method = http.MethodDelete
)

func (m Method) IsOk() bool {
	switch m {
	case MethodGet, MethodPost, MethodPut, MethodPatch, MethodDelete:
		return true
	}
	return false
}

func (m Method) String() string {
	return string(m)
}

type Schema map[Route][]Method

func (s Schema) Match(method, route string) bool {
	if len(s) == 0 {
		return false
	}
	mm := Method(method)
	if !mm.IsOk() {
		return false
	}
	ms, ok := s[route]
	if !ok {
		return false
	}
	return slices.Contains(ms, mm)
}
