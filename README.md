### CAPYBARAS ARE LOVE

[![Build Status](https://travis-ci.org/monkeydioude/capybara.svg?branch=master)](https://travis-ci.org/monkeydioude/capybara)

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

Entries in "services" by id:
- "duck" will redirect any request caught on port 80, starting with "/kwak" to the port 9090 using a **regex** as matching (default) method.
- "stoned" will redirect any request caught on port 80, starting with "/jesus" to the port 9091 using a **string** as matching method. This method compare the string with the beginning of the URI. It does not try to find the string inside the URI.

**/ ! \ On Linux (did not try on other system), capybara must be run with sudo if chosen proxy port is under 1024. It will fail otherwise.**


This project still needs:
- Tests
- Refacto (so it can be more easily tested)
- Allow config refresh through config file
- More methods of matching besides "string" and "regex"
- Refacto "RemovePattern" behavior to a "RedirectPath" behavior


This project might need:
- Better logging ?
