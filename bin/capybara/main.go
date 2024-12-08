package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/monkeydioude/capybara/internal/capybara"
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
}

func handleTLS(conf *capybara.Config, server *http.Server, handler *capybara.Handler) func() error {
	if conf.Proxy.TLSHost == "" {
		return func() error {
			return server.ListenAndServe()
		}
	}
	if conf.Proxy.TLSHost == "localhost" {
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
		Email:      "monkeydioude@gmail.com",
		HostPolicy: autocert.HostWhitelist(conf.Proxy.TLSHost), //Your domain here
		Cache:      autocert.DirCache(cacheDir),                //Folder for storing certificates
	}
	server.TLSConfig = certManager.TLSConfig()
	cert, err := server.TLSConfig.GetCertificate(&tls.ClientHelloInfo{ServerName: conf.Proxy.TLSHost})
	if err != nil {
		log.Fatalf("could not retrieve any cert: %s", err)
	}

	handler.SetCredentials(credentials.NewServerTLSFromCert(cert))
	server.Addr = ":https"

	return func() error {
		go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
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
	go capybara.UpdateServicesRoutine(handler, *cp, capybara.UPDATE_SERVICE_TIMER)
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", conf.Proxy.Port),
		Handler:        handler,
		IdleTimeout:    120 * time.Second, // Prevent idle connections from closing prematurely
		MaxHeaderBytes: 1 << 20,
	}
	serve := handleTLS(conf, server, handler)
	startingLog(conf)
	if err := serve(); err != nil {
		log.Fatal(err)
	}
}
