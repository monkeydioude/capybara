package capybara

import (
	"log/slog"
	"net/http"

	"github.com/CAFxX/httpcompression"
)

// compressibleTypes are the responses worth compressing: text, and the code and data of web
// pages. Images, fonts, archives and videos are compressed already.
var compressibleTypes = []string{
	"text/html", "text/plain", "text/css", "text/csv", "text/markdown", "text/xml",
	"text/javascript", "application/javascript",
	"application/json", "application/problem+json", "application/ld+json", "application/manifest+json", "application/x-ndjson",
	"application/xml", "image/svg+xml",
}

// compressMinSize: smaller responses gain less than the compression costs.
const compressMinSize = 1024

// Compression wraps an HTTP handler so its responses go out compressed, brotli or else gzip,
// as the client accepts. A response compressed already (its own Content-Encoding) goes out as
// it is, and so does any response to a Range request (the range is of the uncompressed bytes)
// or to a protocol upgrade (a WebSocket behind a service set up as plain http).
type Compression func(http.Handler) http.Handler

// NewCompression is the compression capybara applies to HTTP services. Should it fail to set
// up, responses go out uncompressed: a slower gateway rather than none.
func NewCompression() Compression {
	compress, err := httpcompression.DefaultAdapter(
		httpcompression.ContentTypes(compressibleTypes, false),
		httpcompression.MinSize(compressMinSize),
		httpcompression.BrotliCompressionLevel(5), // most of brotli's gain, fast enough per request
	)
	if err != nil {
		slog.Error("Could not set compression up: responses go out uncompressed", "error", err.Error())
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		compressed := compress(next)
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// the library would drop the Range header of every request: keep partial downloads working
			if r.Header.Get("Range") != "" || r.Header.Get("Upgrade") != "" {
				next.ServeHTTP(rw, r)
				return
			}
			compressed.ServeHTTP(rw, r)
		})
	}
}
