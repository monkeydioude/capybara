package main

import (
	"net/http"
	"slices"
)

var protectedCerts = map[string][2]string{
	"localhost":    {"certs/localhost.crt", "certs/localhost.key"},
	"test.dev":     {"certs/test.dev.crt", "certs/test.dev.key"},
	"api.test.dev": {"certs/test.dev.crt", "certs/test.dev.key"},
}

func checkProtectedCerts(certHosts []string, server *http.Server) func() error {
	for label, certFiles := range protectedCerts {
		if slices.Contains(certHosts, label) {
			return func() error {
				return server.ListenAndServeTLS(certFiles[0], certFiles[1])
			}
		}
	}
	return nil
}
