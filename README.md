### CAPYBARAS ARE LOVE

[![Build Status](https://travis-ci.org/monkeydioude/capybara.svg?branch=master)](https://travis-ci.org/monkeydioude/capybara)

Simple Reverse-Proxy written in go.

Listen to a port and redirect any url matching a pattern to another port.

I made this as a solution for hosting multiple services on the same "url" without relying on having a port straight in the URL. Some companies/organizations block any request to a url using a port as one of their security measures.

**First parameter must be the path to your config json file**

Example of config.json

```
{
    "proxy": {
        "port": 80
    },
    "services": [
        {
            "id": "snoopdorkydork",
            "pattern": "^/kwak/",
            "port": 9090
        }
    ]    
}

```

This config will redirect any request caught on port 80, starting with "/kwak" to the port 9090.

**/ ! \ On Linux (did not try on other system), capybara must be run with sudo if chosen proxy port is under 1024. It will fail otherwise.**
