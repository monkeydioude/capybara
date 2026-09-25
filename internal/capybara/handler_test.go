package capybara

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dh *DummyHTTPHandler
var dr *DummyResponseWriter

func TestIcanBuildURL(t *testing.T) {
	if defaultLocalhost+":9090" != buildURL(9090) {
		t.Fail()
	}
}

func TestOnlyPublicRoutesAreServed(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}))
	defer backend.Close()
	u, _ := url.Parse(backend.URL)
	port, _ := strconv.Atoi(u.Port())

	c := NewConfig("../../testdata/routes.config.yaml")
	c.Services[0].Port = int32(port)
	h := NewHandler(c.Services)

	for _, tc := range []struct {
		method string
		target string
		code   int
	}{
		// users: [GET, POST, PATCH]
		{http.MethodGet, "/bypasscors/users", http.StatusOK},
		{http.MethodPost, "/bypasscors/users", http.StatusOK},
		{http.MethodPatch, "/bypasscors/users", http.StatusOK},
		{http.MethodGet, "/bypasscors/users?page=2", http.StatusOK},
		{http.MethodDelete, "/bypasscors/users", http.StatusNotFound},
		{http.MethodGet, "/bypasscors/users/42", http.StatusNotFound},
		{http.MethodGet, "/users", http.StatusNotFound},
		// books: ["*"]
		{http.MethodGet, "/bypasscors/books", http.StatusOK},
		{http.MethodPost, "/bypasscors/books", http.StatusOK},
		{http.MethodPut, "/bypasscors/books", http.StatusOK},
		{http.MethodPatch, "/bypasscors/books", http.StatusOK},
		{http.MethodDelete, "/bypasscors/books", http.StatusOK},
		{http.MethodOptions, "/bypasscors/books", http.StatusOK},
		// books*: [GET]
		{http.MethodGet, "/bypasscors/books/42", http.StatusOK},
		{http.MethodGet, "/bypasscors/books/42/pages", http.StatusOK},
		{http.MethodPost, "/bypasscors/books/42", http.StatusNotFound},
		{http.MethodGet, "/bypasscors/booksellers", http.StatusNotFound},
		{http.MethodGet, "/bypasscors/books/../admin", http.StatusNotFound},
		{http.MethodGet, "/bypasscors/books/%2e%2e/admin", http.StatusNotFound},
		// not listed
		{http.MethodGet, "/bypasscors", http.StatusNotFound},
		{http.MethodGet, "/bypasscors/admin", http.StatusNotFound},
	} {
		r := httptest.NewRequest(tc.method, tc.target, nil)
		r.Host = "spendbaker.com"
		rw := httptest.NewRecorder()

		h.ServeHTTP(rw, r)

		assert.Equal(t, tc.code, rw.Code, "%s %s", tc.method, tc.target)
	}
}

func init() {
	dh = &DummyHTTPHandler{}
	dr = &DummyResponseWriter{}
}
