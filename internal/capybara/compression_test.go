package capybara

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var bigJSON = `{"cards":[` + strings.Repeat(`{"title":"a card","tags":["bug","urgent"]},`, 200) + `{}]}`

// upstream answers body with contentType, and the extra headers; it records the Range it got.
func upstream(body, contentType string, extra map[string]string, gotRange *string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if gotRange != nil {
			*gotRange = r.Header.Get("Range")
		}
		rw.Header().Set("Content-Type", contentType)
		for k, v := range extra {
			rw.Header().Set(k, v)
		}
		_, _ = io.WriteString(rw, body)
	})
}

func serve(t *testing.T, h http.Handler, headers map[string]string) *http.Response {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/api/cards", nil)
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	rw := httptest.NewRecorder()
	NewCompression()(h).ServeHTTP(rw, r)
	return rw.Result()
}

func decoded(t *testing.T, res *http.Response) string {
	t.Helper()
	var reader io.Reader = res.Body
	switch res.Header.Get("Content-Encoding") {
	case "br":
		reader = brotli.NewReader(res.Body)
	case "gzip":
		gz, err := gzip.NewReader(res.Body)
		require.NoError(t, err)
		reader = gz
	}
	b, err := io.ReadAll(reader)
	require.NoError(t, err)
	return string(b)
}

func TestCompressionPicksBrotliThenGzip(t *testing.T) {
	res := serve(t, upstream(bigJSON, "application/json; charset=utf-8", nil, nil), map[string]string{"Accept-Encoding": "gzip, deflate, br"})
	assert.Equal(t, "br", res.Header.Get("Content-Encoding"))
	assert.Contains(t, res.Header.Values("Vary"), "Accept-Encoding")
	assert.Equal(t, bigJSON, decoded(t, res))

	res = serve(t, upstream(bigJSON, "application/json", nil, nil), map[string]string{"Accept-Encoding": "gzip"})
	assert.Equal(t, "gzip", res.Header.Get("Content-Encoding"))
	assert.Equal(t, bigJSON, decoded(t, res))
}

func TestCompressionLeavesAlone(t *testing.T) {
	br := map[string]string{"Accept-Encoding": "br, gzip"}
	cases := map[string]struct {
		handler http.Handler
		headers map[string]string
	}{
		"compressed already":  {upstream("precompressed bytes", "text/javascript", map[string]string{"Content-Encoding": "br"}, nil), br},
		"an image":            {upstream(bigJSON, "image/png", nil, nil), br},
		"a small response":    {upstream(`{"ok":true}`, "application/json", nil, nil), br},
		"no accepted codings": {upstream(bigJSON, "application/json", nil, nil), map[string]string{}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			res := serve(t, c.handler, c.headers)
			got, _ := io.ReadAll(res.Body)
			if name == "compressed already" {
				assert.Equal(t, "br", res.Header.Get("Content-Encoding"), "its own encoding")
				assert.Equal(t, "precompressed bytes", string(got))
				return
			}
			assert.Empty(t, res.Header.Get("Content-Encoding"))
			assert.NotEmpty(t, got)
		})
	}
}

func TestCompressionSkipsUpgrades(t *testing.T) {
	res := serve(t, upstream(bigJSON, "application/json", nil, nil), map[string]string{"Accept-Encoding": "br", "Connection": "Upgrade", "Upgrade": "websocket"})
	assert.Empty(t, res.Header.Get("Content-Encoding"))
}

func TestCompressionKeepsRangeRequests(t *testing.T) {
	var gotRange string
	res := serve(t, upstream(bigJSON, "application/json", nil, &gotRange), map[string]string{"Accept-Encoding": "br", "Range": "bytes=0-99"})
	assert.Equal(t, "bytes=0-99", gotRange, "the upstream still gets the range")
	assert.Empty(t, res.Header.Get("Content-Encoding"))
}
