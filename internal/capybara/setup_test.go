package capybara

import (
	"net/http"
)

// This file is only here to setup tools for testing purpose

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
