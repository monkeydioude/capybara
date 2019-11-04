package capybara

import (
	"net/http"
	"testing"
)

// This file is only here to setup tools for testing purpose

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func assertNotEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Fatalf("%s != %s", a, b)
	}
}

type DummyHTTPHandler struct {
}

func (d *DummyHTTPHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
}

type DummyResponseWriter struct {
	write func([]byte) (int, error)
}

func (dr *DummyResponseWriter) Header() http.Header {
	return nil
}

func (dr *DummyResponseWriter) Write(b []byte) (int, error) {
	return dr.write(b)
}

func (dr *DummyResponseWriter) WriteHeader(statusCode int) {
}
