package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type proxy struct {
	Port int `json:"port"`
}

type service struct {
	ID            string `json:"id"`
	Pattern       string `json:"pattern"`
	Port          int    `json:"port"`
	RemovePattern bool   `json:"removePattern,omitempty"`
}

type config struct {
	Proxy    proxy      `json:"proxy"`
	Services *[]service `json:"services"`
}

type handler struct {
	s *[]service
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/favicon.ico" {
		rw.WriteHeader(200)
		return
	}

	for _, service := range *h.s {
		if _, err := matchAndFind(service.Pattern, r.RequestURI); err != nil {
			continue
		}

		u, err := url.Parse(fmt.Sprintf("http://localhost:%d", service.Port))
		if err != nil {
			log.Printf("[WARN] Could not parse url %s", fmt.Sprintf("http://localhost:%d", service.Port))
			continue
		}

		rp := httputil.NewSingleHostReverseProxy(u)
		if service.RemovePattern {
			r.URL.Path = "/"
		}
		rp.ServeHTTP(rw, r)
		return
	}

	http.NotFound(rw, r)
}

func newConfig(p string) (c *config) {
	d, err := ioutil.ReadFile(p)

	if err != nil {
		log.Fatalf("[ERR ] Could not ReadFile, reason: %s", err)
	}

	err = json.Unmarshal(d, &c)

	if err != nil {
		log.Fatalf("[ERR ] Could not Unmarshal config, reason: %s", err)
	}

	return
}

func main() {
	cp := flag.String("c", "", "Path to config file")
	flag.Parse()
	if *cp == "" {
		log.Fatal("[ERR ] Path to Config json file is required.")
	}

	c := newConfig(*cp)

	handler := &handler{
		s: c.Services,
	}

	go updateServicesRoutine(handler, *cp)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", c.Proxy.Port),
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("[INFO] Starting server")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
