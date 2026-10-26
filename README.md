### CAPYBARAS ARE LOVE

[![Go](https://github.com/monkeydioude/capybara/actions/workflows/go.yml/badge.svg)](https://github.com/monkeydioude/capybara/actions/workflows/go.yml)

Simple Reverse-Proxy written in go.

Listen to a port and redirect any url matching a pattern to another port.

I made this as a solution for hosting multiple services on the same "url" without relying on having a port straight in the URL. Some companies/organizations block any request to a url using a port as one of their security measures.

**First parameter must be the path to your json/yaml config file**

Example of config.json

```
{
    "proxy": {
        "port": 80
    },
    "services": [
        {
            "id": "duck",
            "pattern": "^/kwak/",
            "port": 9090
        },
        {
            "id": "stoned",
            "method": "string",
            "pattern": "/jesus",
            "port": 9091
        }
    ]
}

```

Example of config.yaml

```
{
    proxy:
        port: 80
        tls_hosts:
            - localhost

    services:
        -   id: bypasscors
            method: string
            pattern: /bypasscors
            port: 8080

}

```

Entries in "services" by id:

- "duck" will redirect any request caught on port 80, starting with "/kwak" to the port 9090 using a **regex** as matching (default) method.
- "stoned" will redirect any request caught on port 80, starting with "/jesus" to the port 9091 using a **string** as matching method. This method compare the string with the beginning of the URI. It does not try to find the string inside the URI.

#### Public routes

A service can restrict which routes are publicly reachable with the "routes" key. Without it, every route matching the pattern is served.

```
services:
    -   id: bypasscors
        method: string
        pattern: /bypasscors
        port: 8080
        routes:
            users: [GET, POST, PATCH]
            books: ["*"]
            books*: [GET]
```

- Routes are relative to the pattern: `users` matches `/bypasscors/users`. The query string is ignored.
- `users` only matches `users`, not `users/42`.
- A trailing `*` matches the route and everything under it: `books*` matches `books`, `books/42` and `books/42/pages`, but not `booksellers`. `books/*` only matches what's under `books/`, and `*` alone matches every route.
- `"*"` as a method allows any method. **It must be quoted in YAML**: an unquoted `*` is a YAML alias and the config won't load.
- When several routes match, the most specific one decides: an exact route beats a wildcard, and a longer wildcard beats a shorter one. Above, `POST /bypasscors/books` is allowed by `books`, while `POST /bypasscors/books/42` is refused by `books*`.
- Paths containing `.` or `..` segments are always refused.
- Anything else gets a 404. Methods aren't implied: list `OPTIONS` if browsers need CORS preflight on a route.

### Compression

Responses of HTTP services go out compressed, brotli or else gzip, as the client accepts (`Accept-Encoding`), with `Vary: Accept-Encoding`. Only text-like types (HTML, CSS, JS, JSON, XML, SVG, CSV, Markdown, plain text) of 1 kB or more. Left as they are:

- responses the service compressed itself (they have their own `Content-Encoding`)
- responses to `Range` requests (partial downloads, video seeking)
- WebSocket and gRPC services

### Operations

- `capybara -check -c config.yaml` validates a config file and exits 1 if it would be refused. A config is refused when it has no services, an entry is empty, a `pattern` is missing (an empty pattern would match every request), a regex pattern does not compile, a `port` is outside 1-65535, or a `method`/`protocol` is unknown. At startup a refused config is fatal; on the 30 s reload the error is logged and the current services stay in place. Run `-check` on the production file before replacing it.
- Timeouts: 10 s to finish the TLS handshake and send the request headers, 120 s idle. Request bodies and responses are not bounded, so uploads, WebSocket and gRPC streams are not cut. Backends: 5 s to connect, 60 s to start answering, at most 1024 connections per backend. At most 10000 concurrent client connections per listener.
- `CONNECT` and `TRACE` are refused with 405. `/_healthcheck` and `/favicon.ico` answer `GET` and `HEAD` only.
- `SIGINT`/`SIGTERM` stop the listeners and let in-flight requests finish for up to 10 s. The process exits 1 when a listener fails.
- Build with the Go toolchain pinned in `go.mod`: older toolchains have known remotely triggerable crash bugs in the HTTP/2 and TLS stack. CI runs `govulncheck`.

**/ ! \ On Linux (did not try on other system), capybara must be run with sudo if chosen proxy port is under 1024. It will fail otherwise.**

This project still needs:

- Tests
- Refacto (so it can be more easily tested)
- Allow config refresh through config file
- More methods of matching besides "string" and "regex"
- Refacto "RemovePattern" behavior to a "RedirectPath" behavior

This project might need:

- Better logging ?
