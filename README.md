# Prometheus metrics module for Caddy v2

This implements a [Caddy v2 module](https://caddyserver.com/docs/extending-caddy)
that exposes metrics in the [Prometheus](https://prometheus.io/) format.

## Build caddy with caddyprom module
Download xcaddy latest relsease : https://github.com/caddyserver/xcaddy/releases/  
build caddy with prometheus : 
```console
./xcaddy build --with https://github.com/hairyhenderson/caddyprom
```

## Usage
### Caddyfile
The simplest use could be in a Caddyfile like:

```
{
    order prometheus first
}

localhost

prometheus
```

### Json config
Here is an exemple using caddy as a reverse proxy
```javascript
{
    "apps": {
        "http": {
            "servers": {
                "srv0": {
                    "listen": [
                        ":443"
                    ],
                    "routes": [
                        {
                            "handle": [
                                {
                                    "handler": "subroute",
                                    "routes": [
                                        {
                                            "handle": [
                                                {
                                                    "handler": "prometheus"
                                                },
                                                {
                                                    "handler": "reverse_proxy",
                                                    "upstreams": [
                                                        {
                                                            "dial": "10.0.0.1:80"
                                                        },
                                                        {
                                                            "dial": "10.0.0.2:80"
                                                        }
                                                    ]
                                                }
                                            ]
                                        }
                                    ]
                                }
                            ],
                            "match": [
                                {
                                    "host": [
                                        "redacted.mycompany.com"
                                    ]
                                }
                            ],
                            "terminal": true
                        }
                    ]
                }
            }
        }
    }
}
```

## Get metrics
Then, when using a Caddy server with this module enabled:

```console
$ curl localhost/
$ curl localhost:9180/metrics
...
caddy_http_response_size_bytes_sum{code="418",method="get"} 42
...
```

## License

[The MIT License](http://opensource.org/licenses/MIT)

Copyright (c) 2020 Dave Henderson
