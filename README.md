# Prometheus metrics module for Caddy v2

This implements a [Caddy v2 module](https://caddyserver.com/docs/extending-caddy)
that exposes metrics in the [Prometheus](https://prometheus.io/) format.

## Usage

The simplest use could be in a Caddyfile like:

```
{
    order prometheus first
}

localhost

prometheus
```

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
