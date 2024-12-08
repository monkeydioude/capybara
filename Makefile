DEFAULT_CERT_PATH=bin/capybara/certs
CERT_PATH ?= ${DEFAULT_CERT_PATH}

.PHONY: certs
certs:
	openssl req -newkey rsa:2048 -nodes -keyout ${CERT_PATH}/localhost.key -x509 -days 365 -out ${CERT_PATH}/localhost.crt -subj "/CN=localhost"