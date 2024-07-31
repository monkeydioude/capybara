package capybara

import (
	"net/http"
)

func HealthcheckHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(200)
	rw.Write([]byte("{\"health\": \"OK\"}"))
}
