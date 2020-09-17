package caddyprom

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

const (
	defaultPath = "/metrics"
	defaultAddr = "0.0.0.0:9180"
)

var (
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
	responseStatus  *prometheus.CounterVec
	responseLatency *prometheus.HistogramVec
)

func (m *Metrics) initMetrics(ctx caddy.Context) error {
	log := ctx.Logger(m)

	m.registerMetrics("caddy", "http")
	if m.Path == "" {
		m.Path = defaultPath
	}
	if m.Addr == "" {
		m.Addr = defaultAddr
	}

	if !m.useCaddyAddr {
		mux := http.NewServeMux()
		mux.Handle(m.Path, m.metricsHandler)

		srv := &http.Server{Handler: mux}
		// if m.Addr does not have a port just add the default one
		if !strings.Contains(m.Addr, ":") {
			m.Addr += ":" + strings.Split(defaultAddr, ":")[1]
		}
		zap.S().Info("Binding prometheus exporter to %s", m.Addr)
		listener, err := net.Listen("tcp", m.Addr)
		if err != nil {
			return fmt.Errorf("failed to listen to %s: %w", m.Addr, err)
		}

		go func() {
			err := srv.Serve(listener)
			if err != nil && err != http.ErrServerClosed {
				log.Error("metrics handler's server failed to serve", zap.Error(err))
			}
		}()
	}
	return nil
}

func (m *Metrics) registerMetrics(namespace, subsystem string) {
	if m.latencyBuckets == nil {
		m.latencyBuckets = []float64{.05, .2, 1, 3, 10, 60}
	}
	if m.sizeBuckets == nil {
		m.sizeBuckets = []float64{0, 500, 5000, 50000, 500000, 5000000}
	}

	// TODO: add "handler" and probably others
	httpLabels := []string{"code", "method", "path"}

	requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_count_total",
		Help:      "Counter of HTTP(S) requests made.",
	}, httpLabels)

	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_duration_seconds",
		Help:      "Histogram of the time (in seconds) each request took.",
		Buckets:   m.latencyBuckets,
	}, httpLabels)

	responseSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "response_size_bytes",
		Help:      "Size of the returns response in bytes.",
		Buckets:   m.sizeBuckets,
	}, httpLabels)

	responseStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "response_status_count_total",
		Help:      "Counter of response status codes.",
	}, httpLabels)

	// TODO: I guess this should be time-to-first-byte?
	// responseLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
	// 	Namespace: namespace,
	// 	Subsystem: subsystem,
	// 	Name:      "response_latency_seconds",
	// 	Help:      "Histogram of the time (in seconds) until the first write for each request.",
	// 	Buckets:   m.latencyBuckets,
	// }, httpLabels)
}
