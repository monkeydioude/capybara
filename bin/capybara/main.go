package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/monkeydioude/capybara/pkg/capybara"
	"golang.org/x/crypto/acme/autocert"
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
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serve := func() error {
		return server.ListenAndServe()
	}

	if conf.Proxy.TLSHost != "" {
		cacheDir := "certs"
		if conf.Proxy.TLSHost != "localhost" {
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
			handler.SetCertificate(cert)
			server.Addr = ":https"
			// handler
			serve = func() error {
				go http.ListenAndServe(":http", certManager.HTTPHandler(nil))

				return server.ListenAndServeTLS("", "")
			}
		} else {
			cert, err := tls.LoadX509KeyPair("certs/cert.pem", "certs/key.pem")
			if err != nil {
				log.Fatalf("could not retrieve any localhost cert: %s", err)
			}
			handler.SetCertificate(&cert)
			serve = func() error {

				return server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem")
			}
		}
	}

	startingLog(conf)
	if err := serve(); err != nil {
		log.Fatal(err)
	}
}
