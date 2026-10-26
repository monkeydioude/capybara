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
	maxHeaderBytes    = 64 << 10
	// maxConns caps the concurrent client connections of a listener: past it, accepts wait
	// instead of the process running out of file descriptors.
	maxConns        = 10000
	acmeMaxConns    = 2000
	shutdownTimeout = 10 * time.Second
)

func newHTTP2Config() *http.HTTP2Config {
	return &http.HTTP2Config{
		MaxConcurrentStreams: 100,
		PingTimeout:          15 * time.Second,
		SendPingTimeout:      30 * time.Second,
		WriteByteTimeout:     30 * time.Second,
	}
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
		ln, err := listenCapped(server.Addr, maxConns)
		if err != nil {
			return err
		}
		return server.Serve(ln)
	}
}

func serveTLS(server *http.Server, certFile, keyFile string) func() error {
	return func() error {
		ln, err := listenCapped(server.Addr, maxConns)
		if err != nil {
			return err
		}
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
	g.Add(func() error {
		ln, err := listenCapped(acmeServer.Addr, acmeMaxConns)
		if err != nil {
			return err
		}
		return acmeServer.Serve(ln)
	}, shutdown(acmeServer))
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
	serveFn := handleTLS(g, conf, server, handler)
	g.Add(func() error {
		startingLog(conf)
		return serveFn()
	}, shutdown(server))
	err := g.Run()
	var sig run.SignalError
	if errors.As(err, &sig) {
		slog.Info("Stopping capybara", "signal", sig.Signal)
		return
	}
	if err != nil {
		slog.Error("Stopping capybara", "error", err)
		os.Exit(1)
	}
}
