package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/monkeydioude/capybara/internal/capybara"
	"github.com/oklog/run"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc/credentials"
)

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

func handleTLS(
	g *run.Group,
	conf *capybara.Config,
	server *http.Server,
	handler *capybara.Handler,
) func() error {
	var certHosts []string
	if conf.Proxy.TLSHost != "" {
		certHosts = append(conf.Proxy.TLSHosts, conf.Proxy.TLSHost)
	}
	if len(certHosts) == 0 {
		return func() error {
			return http.ListenAndServe(fmt.Sprintf(":%d", conf.Proxy.Port), handler)
		}
	}
	proCertFn := checkProtectedCerts(certHosts, server)
	if proCertFn != nil {
		server.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		handler.SetCredentials(credentials.NewTLS(server.TLSConfig))

		return func() error {
			return server.ListenAndServeTLS("certs/localhost.crt", "certs/localhost.key")
		}
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
	g.Add(func() error {
		return http.ListenAndServe(":http", certManager.HTTPHandler(nil))
	}, func(err error) {
		slog.Error("Stoping http server", "error", err)
	})
	server.TLSConfig = certManager.TLSConfig()
	server.Addr = ":https"

	return func() error {
		return server.ListenAndServeTLS("", "")
	}
}

func main() {
	cp := flag.String("c", "", "Path to config file")
	flag.Parse()
	if *cp == "" {
		log.Fatal("[ERR ] Path to Config json file is required.")
	}

	conf := capybara.NewConfig(*cp)
	handler := capybara.NewHandler(conf.Services)
	g := &run.Group{}
	g.Add(func() error {
		capybara.UpdateServicesRoutine(handler, *cp, capybara.UPDATE_SERVICE_TIMER)
		return nil
	}, func(err error) {
		slog.Error("Stoping updating services", "error", err)
	})
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", conf.Proxy.Port),
		Handler:        handler,
		IdleTimeout:    120 * time.Second, // Prevent idle connections from closing prematurely
		MaxHeaderBytes: 1 << 20,
	}
	serve := handleTLS(g, conf, server, handler)
	g.Add(func() error {
		startingLog(conf)
		return serve()
	}, func(err error) {
		slog.Error("Stoping server", "error", err)
	})
	if err := g.Run(); err != nil {
		slog.Error("Stoping capybara", "error", err)
	}
}
