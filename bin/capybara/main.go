package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/monkeydioude/capybara/internal/capybara"
	"github.com/oklog/run"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc/credentials"
)

const (
	// readHeaderTimeout bounds the TLS handshake and the reading of the request line and
	// headers: the only phase a client can stretch without sending anything. There is no
	// ReadTimeout nor WriteTimeout on purpose, they would cut uploads, downloads, WebSocket
	// and gRPC streams.
	readHeaderTimeout = 10 * time.Second
	idleTimeout       = 120 * time.Second
	maxHeaderBytes    = 1 << 20
	shutdownTimeout   = 10 * time.Second
)

func newHTTP2Config() *http.HTTP2Config {
	return &http.HTTP2Config{
		PingTimeout:      15 * time.Second,
		SendPingTimeout:  30 * time.Second,
		WriteByteTimeout: 30 * time.Second,
	}
}

// connLimit caps the concurrent connections of a listener at 1/share of the file
// descriptor budget, so that a connection flood degrades into waiting accepts instead of
// the process failing for lack of descriptors. Backend connections and the other listener
// need the rest of the budget.
func connLimit(share int) int {
	var rl syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rl); err != nil || rl.Cur == 0 {
		return 4096
	}
	budget := rl.Cur
	if budget > 1<<30 {
		budget = 1 << 30
	}
	return max(int(budget)/share, 1024)
}

func startingLog(conf *capybara.Config) {
	log.Printf("[INFO] Starting server on port :%d\n", conf.Proxy.Port)
	var b strings.Builder

	for _, s := range conf.Services {
		b.WriteString("\t - ")
		b.WriteString(s.String())
		b.WriteString("\n")
	}
	log.Printf("[INFO] Available redirection services:\n %s", b.String())
	log.Printf("[INFO] Healthcheck available at /_healthcheck")
}

// listenCapped listens on addr with a cap on concurrent connections.
func listenCapped(addr string, limit int) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return netutil.LimitListener(ln, limit), nil
}

func serve(server *http.Server) func() error {
	return func() error {
		limit := connLimit(4)
		ln, err := listenCapped(server.Addr, limit)
		if err != nil {
			return err
		}
		slog.Info("Listening", "addr", server.Addr, "max_connections", limit)
		return server.Serve(ln)
	}
}

func serveTLS(server *http.Server, certFile, keyFile string) func() error {
	return func() error {
		limit := connLimit(4)
		ln, err := listenCapped(server.Addr, limit)
		if err != nil {
			return err
		}
		slog.Info("Listening", "addr", server.Addr, "max_connections", limit)
		return server.ServeTLS(ln, certFile, keyFile)
	}
}

// shutdown is the run.Group interrupt of a server: it stops accepting and lets in-flight
// requests finish for up to shutdownTimeout.
func shutdown(server *http.Server) func(error) {
	return func(error) {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Stopping server", "addr", server.Addr, "error", err)
		}
	}
}

func handleTLS(
	ctx context.Context,
	g *run.Group,
	conf *capybara.Config,
	server *http.Server,
	handler *capybara.Handler,
) func() error {
	certHosts := conf.Proxy.TLSHosts
	if conf.Proxy.TLSHost != "" {
		certHosts = append(certHosts, conf.Proxy.TLSHost)
	}
	if len(certHosts) == 0 {
		return serve(server)
	}
	if checkProtectedCerts(certHosts, server) != nil {
		server.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		handler.SetCredentials(credentials.NewTLS(server.TLSConfig))

		return serveTLS(server, "certs/localhost.crt", "certs/localhost.key")
	}
	cacheDir := "certs"
	if conf.Proxy.TLSCacheDir != "" {
		cacheDir = conf.Proxy.TLSCacheDir
	}
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Email:      conf.Proxy.Email,
		HostPolicy: autocert.HostWhitelist(certHosts...),
		Cache:      autocert.DirCache(cacheDir),
	}
	// The :80 listener only answers ACME challenges and redirects to HTTPS: no request
	// carries a body, so full timeouts are safe here.
	acmeServer := &http.Server{
		Addr:              ":http",
		Handler:           certManager.HTTPHandler(nil),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	// :80 is a convenience (HTTP-01 challenges and the redirect): without it autocert still
	// issues certificates over :443 (TLS-ALPN-01), so a failed bind must not stop the proxy.
	g.Add(func() error {
		limit := connLimit(16)
		ln, err := listenCapped(acmeServer.Addr, limit)
		if err != nil {
			slog.Error("Could not listen on :80, continuing without the ACME/redirect listener", "error", err)
			<-ctx.Done()
			return nil
		}
		slog.Info("Listening", "addr", acmeServer.Addr, "max_connections", limit)
		return acmeServer.Serve(ln)
	}, func(error) {
		// nothing to drain on :80: close it at once so that :443 starts draining right away
		if err := acmeServer.Close(); err != nil {
			slog.Error("Stopping server", "addr", acmeServer.Addr, "error", err)
		}
	})
	server.TLSConfig = certManager.TLSConfig()
	server.Addr = ":https"

	return serveTLS(server, "", "")
}

func main() {
	cp := flag.String("c", "", "Path to config file")
	check := flag.Bool("check", false, "Validate the config file and exit")
	flag.Parse()
	if *cp == "" {
		log.Fatal("[ERR ] Path to Config json file is required.")
	}
	if *check {
		if _, err := capybara.LoadConfig(*cp); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("config is valid")
		return
	}

	conf := capybara.NewConfig(*cp)
	handler := capybara.NewHandler(conf.Services)
	g := &run.Group{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g.Add(run.SignalHandler(ctx, os.Interrupt, syscall.SIGTERM))
	g.Add(func() error {
		capybara.UpdateServicesRoutine(ctx, handler, *cp, capybara.UPDATE_SERVICE_TIMER)
		return nil
	}, func(error) {
		cancel()
	})
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", conf.Proxy.Port),
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
		HTTP2:             newHTTP2Config(),
	}
	serveFn := handleTLS(ctx, g, conf, server, handler)
	g.Add(func() error {
		startingLog(conf)
		return serveFn()
	}, shutdown(server))
	err := g.Run()
	if errors.Is(err, run.ErrSignal) {
		slog.Info("Stopping capybara", "reason", err.Error())
		return
	}
	if err != nil {
		slog.Error("Stopping capybara", "error", err)
		os.Exit(1)
	}
}
