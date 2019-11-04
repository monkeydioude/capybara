package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/monkeydioude/capybara/pkg/capybara"
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

	startingLog(conf)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
