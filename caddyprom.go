// Package caddyprom implements a metrics module for Caddy v2 that exports in
// the Prometheus text format.
//
// The simplest use could be in a Caddyfile like:
//
//    {
//        order prometheus first
//    }
//    localhost
//
//    prometheus
//
package caddyprom // import "github.com/hairyhenderson/caddyprom"

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Metrics{})
	httpcaddyfile.RegisterHandlerDirective("prometheus", parseCaddyfile)
}

// Metrics -
type Metrics struct {
	Addr string `json:"address,omitempty"`
	Path string	`json:"path,omitempty"`

	useCaddyAddr   bool
	latencyBuckets []float64
	sizeBuckets    []float64
	metricsHandler http.Handler
}

// CaddyModule returns the Caddy module information.
func (Metrics) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.prometheus",
		New: func() caddy.Module { return new(Metrics) },
	}
}

type zapLogger struct {
	zl *zap.Logger
}

func (l *zapLogger) Println(v ...interface{}) {
	l.zl.Sugar().Error(v...)
}

// Provision -
func (m *Metrics) Provision(ctx caddy.Context) error {
	log := ctx.Logger(m)
	m.metricsHandler = promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
		ErrorLog:      &zapLogger{log},
	})

	return m.initMetrics(ctx)
}

// UnmarshalCaddyfile -
func (m *Metrics) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		setInline := d.Args(&m.Addr) // setInline is true if the address was set inline
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "address": // optional: m.Addr has a default value
				if !setInline {
					d.Args(&m.Addr)
				} else {
					return d.Errf("listen address has already been set")
				}
			case "path": // optional: m.Path has a default value
				d.Args(&m.Path)
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
		}
	}
	
	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Metrics
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// ServeHTTP - instrument the handler
// fulfils the caddyhttp.MiddlewareHandler interface
func (m Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) (err error) {
	// TODO: break this out to record the rest of the metrics
	chain :=
		promhttp.InstrumentHandlerCounter(requestCount,
			promhttp.InstrumentHandlerDuration(requestDuration,
				promhttp.InstrumentHandlerResponseSize(responseSize, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					err = next.ServeHTTP(w, r)
				})),
			),
		)

	chain.ServeHTTP(w, r)
	return err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Metrics)(nil)
	_ caddyhttp.MiddlewareHandler = (*Metrics)(nil)
	_ caddyfile.Unmarshaler       = (*Metrics)(nil)
)
