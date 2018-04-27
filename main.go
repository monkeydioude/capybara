package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

type proxy struct {
	Port int `json:"port"`
}

type service struct {
	ID      string `json:"id"`
	Pattern string `json:"pattern"`
	Port    int    `json:"port"`
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

func getConfigPath() (string, error) {
	if len(os.Args) != 2 {
		return "", errors.New("[ERR ] Takes only 1 parameter, config json file path")
	}
	return os.Args[1], nil
}

func main() {
	cp, err := getConfigPath()

	if err != nil {
		log.Fatal(err.Error())
	}
	c := newConfig(cp)

	handler := &handler{
		s: c.Services,
	}

	go updateServicesRoutine(handler, cp)

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
