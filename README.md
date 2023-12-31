# `mon` - A simple local services monitor

`mon` is a simple command-line tool to monitor services. Services are specified in a simple JSON file, containing service name, a URL to check, and optional HTTP headers to send in any requests (for things like specifying Accepts or Authorization headers).

All services are assumed to be HTTP services, and `mon` simply checks that the provided URL for a service returns a `200 (OK)` response.

> `mon` is MacOS-centric currently, and assumes presence of `osascript` for delivering desktop notifications.

## Installation
`mon` is a regular Go program with no third-party dependencies. It can be installed into `$GOBIN` with:

```sh
go install github.com/chriswalker/mon
```

## Configuration
Services to monitor are defined in a simple JSON file, provided to the program with the `--services-file` or `-s` flags.

The file is an array of objects, each of which must minimally define a `name` and `url`. HTTP headers can be specified in the `headers` property.

A sample services file might look like:

```json
[
    { "name": "godocs", "url": "http://localhost:6060" },
    { "name": "rss - miniflux", "url": "http://localhost:8050" },
    {
        "name": "gokrazy",
        "url": "http://gokrazy.local/",
        "headers": {
            "Accepts": "application/json",
            "Authorization": "Basic <bas64 encoded username/pwd here>"
        }
    }
]
```

If a headers map is provided for a service, the specified headers are provided as-is to the request made to the service. Use this for services that might require some kind of authorisation, or where requests must specify what they accept.

## Command-line Flags
| Flag | Description |
| --- | --- |
| `-s`, `--services-file` | Path to the services configuration file (defaults to `./services.json`) |
| `-j`, `--json` | Output status information as JSON (if omitted, defaults to tubular status output) |
| `--notify` | Display a desktop notification for each service that does **not** return a success (`200 OK`) status |

## Automation
A [sample launchd plist file][plist] is provided as a starting point for using `mon` on MacOS. 

[plist]: samples/com.yourdomain.mon.plist
