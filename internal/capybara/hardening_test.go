package capybara

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, path, body string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
}

func loadBody(t *testing.T, body string) (*Config, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfig(t, path, body)
	return LoadConfig(path)
}

func TestLoadConfigRejectsBrokenConfigs(t *testing.T) {
	for name, body := range map[string]string{
		"not json nor yaml":   `{{{`,
		"no services":         `{"services": []}`,
		"null service":        `{"services": [null]}`,
		"no pattern, no host": `{"services": [{"id":"a","port":9000}]}`,
		"bad regex":           `{"services": [{"id":"a","pattern":"^/(","port":9000}]}`,
		"port out of range":   `{"services": [{"id":"a","pattern":"/a","port":0}]}`,
		"unknown method":      `{"services": [{"id":"a","pattern":"/a","port":9000,"method":"glob"}]}`,
		"unknown protocol":    `{"services": [{"id":"a","pattern":"/a","port":9000,"protocol":"ftp"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := loadBody(t, body)
			assert.Error(t, err)
		})
	}
	_, err := LoadConfig(filepath.Join(t.TempDir(), "missing.yaml"))
	assert.Error(t, err)
}

func TestLoadConfigAcceptsWorkingShapes(t *testing.T) {
	for name, body := range map[string]string{
		"host-only catch-all": `{"services": [{"id":"blog","host":"blog.example.com","port":8080}]}`,
		"catch-all pattern":   `{"services": [{"pattern":"/","port":8080}]}`,
		"empty method":        `{"services": [{"pattern":"^/a","method":"","port":8080}]}`,
		"null routes":         `{"services": [{"pattern":"/a","port":8080,"routes":null}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := loadBody(t, body)
			assert.NoError(t, err)
		})
	}
	for _, p := range []string{
		"../../testdata/dummy.config.json",
		"../../testdata/dummy_routine.config.json",
		"../../testdata/routes.config.yaml",
	} {
		_, err := LoadConfig(p)
		assert.NoError(t, err, p)
	}
}

func TestReloadRefusesBrokenConfigAndKeepsServices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfig(t, path, `{"services":[{"id":"one","pattern":"/one","port":9001}]}`)
	c, err := LoadConfig(path)
	require.NoError(t, err)
	h := NewHandler(c.Services)

	for name, body := range map[string]string{
		"empty services":    `{"services": []}`,
		"half-written file": `{"services":[{"id":"tw`,
		"bad entry":         `{"services":[{"id":"two","pattern":"^/(","port":9002}]}`,
	} {
		writeConfig(t, path, body)
		assert.Error(t, reloadServices(h, path), name)
		require.Len(t, h.Services(), 1, name)
		assert.Equal(t, "one", h.Services()[0].ID, name)
	}

	writeConfig(t, path, `{"services":[{"id":"two","pattern":"/two","port":9002}]}`)
	require.NoError(t, reloadServices(h, path))
	assert.Equal(t, "two", h.Services()[0].ID)
}

func TestUpdateServicesRoutineAppliesValidConfigAndStops(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfig(t, path, `{"services":[{"id":"one","pattern":"/one","port":9001}]}`)
	c, err := LoadConfig(path)
	require.NoError(t, err)
	h := NewHandler(c.Services)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		UpdateServicesRoutine(ctx, h, path, 5*time.Millisecond)
		close(done)
	}()

	writeConfig(t, path, `{"services":[{"id":"two","pattern":"/two","port":9002}]}`)
	assert.Eventually(t, func() bool { return h.Services()[0].ID == "two" }, time.Second, 5*time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("UpdateServicesRoutine did not stop on cancel")
	}
}

func TestSetServicesIsSafeDuringRequests(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}))
	defer backend.Close()
	u, _ := url.Parse(backend.URL)
	port, _ := strconv.Atoi(u.Port())
	table := func(id string) []*service {
		return []*service{{ID: id, Pattern: "/api", Method: "string", Port: int32(port)}}
	}
	h := NewHandler(table("a"))
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 200; i++ {
			h.SetServices(table(fmt.Sprint(i)))
		}
	}()
	for i := 0; i < 200; i++ {
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/api/x", nil))
		assert.Equal(t, http.StatusOK, rw.Code)
	}
	<-done
}

func TestConnectAndTraceAreRefused(t *testing.T) {
	h := NewHandler([]*service{{ID: "a", Pattern: "/", Method: "string", Port: 9000}})
	for _, m := range []string{http.MethodConnect, http.MethodTrace} {
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, httptest.NewRequest(m, "/anything", nil))
		assert.Equal(t, http.StatusMethodNotAllowed, rw.Code, m)
		assert.Equal(t, "close", rw.Header().Get("Connection"), m)
	}
}

func TestBuiltinEndpointsAnswerAnyMethodOnTheExactURI(t *testing.T) {
	// no service matches: anything that is not a built-in gets a 404
	h := NewHandler([]*service{{ID: "a", Pattern: "/zzz", Method: "string", Port: 9000}})
	for _, tc := range []struct {
		method, target string
		code           int
	}{
		{http.MethodGet, "/_healthcheck", http.StatusOK},
		{http.MethodHead, "/_healthcheck", http.StatusOK},
		{http.MethodOptions, "/_healthcheck", http.StatusOK}, // HAProxy's default httpchk method
		{http.MethodPost, "/_healthcheck", http.StatusOK},
		{http.MethodGet, "/favicon.ico", http.StatusOK},
		{http.MethodGet, "/favicon.ico?v=2", http.StatusNotFound}, // cache-busted: left to the services
		{http.MethodGet, "/_healthcheck?x=1", http.StatusNotFound},
	} {
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, httptest.NewRequest(tc.method, tc.target, nil))
		assert.Equal(t, tc.code, rw.Code, "%s %s", tc.method, tc.target)
	}
}
