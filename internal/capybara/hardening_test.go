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

func TestLoadConfigRejectsBrokenConfigs(t *testing.T) {
	for name, body := range map[string]string{
		"not json nor yaml": `{{{`,
		"no services":       `{"services": []}`,
		"null service":      `{"services": [null]}`,
		"empty pattern":     `{"services": [{"id":"a","port":9000}]}`,
		"bad regex":         `{"services": [{"id":"a","pattern":"^/(","port":9000}]}`,
		"port out of range": `{"services": [{"id":"a","pattern":"/a","port":0}]}`,
		"unknown method":    `{"services": [{"id":"a","pattern":"/a","port":9000,"method":"glob"}]}`,
		"unknown protocol":  `{"services": [{"id":"a","pattern":"/a","port":9000,"protocol":"ftp"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			writeConfig(t, path, body)
			_, err := LoadConfig(path)
			assert.Error(t, err)
		})
	}
	_, err := LoadConfig(filepath.Join(t.TempDir(), "missing.yaml"))
	assert.Error(t, err)
}

func TestLoadConfigAcceptsTestdata(t *testing.T) {
	for _, p := range []string{
		"../../testdata/dummy.config.json",
		"../../testdata/dummy_routine.config.json",
		"../../testdata/routes.config.yaml",
	} {
		_, err := LoadConfig(p)
		assert.NoError(t, err, p)
	}
}

func TestReloadKeepsLastGoodConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfig(t, path, `{"services":[{"id":"one","pattern":"/one","port":9001}]}`)
	h := NewHandler(NewConfig(path).Services)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go UpdateServicesRoutine(ctx, h, path, 10*time.Millisecond)

	current := func() string { return h.Services()[0].ID }
	writeConfig(t, path, `{"services": []}`) // would blank every route
	time.Sleep(80 * time.Millisecond)
	assert.Equal(t, "one", current(), "an empty config is refused")

	writeConfig(t, path, `{{{`) // a half-written file
	time.Sleep(80 * time.Millisecond)
	assert.Equal(t, "one", current(), "an unparsable config is refused")

	writeConfig(t, path, `{"services":[{"id":"two","pattern":"/two","port":9002}]}`)
	assert.Eventually(t, func() bool { return current() == "two" }, time.Second, 10*time.Millisecond, "a valid config is applied")
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
