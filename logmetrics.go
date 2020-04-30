// Package caddyprom implements a metrics module for Caddy v2 that exports in
// the Prometheus text format.
//
// The simplest use could be in a Caddyfile like:
//
//    localhost
//
//    prometheus
//
package caddyprom // import "github.com/hairyhenderson/caddyprom"

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func init() {
	caddy.RegisterModule(LogMetrics{})
}

// LogMetrics -
type LogMetrics struct {
	Addr string `json:"address,omitempty"`

	useCaddyAddr bool
	// hostname       string
	path           string
	latencyBuckets []float64
	sizeBuckets    []float64
	// subsystem?
	// once sync.Once

	metricsHandler http.Handler

	// the wrapped log encoder
	wrapped zapcore.Encoder

	// The underlying encoder that actually
	// encodes the log entries. Required.
	WrappedRaw json.RawMessage `json:"wrap,omitempty" caddy:"namespace=caddy.logging.encoders inline_key=format"`

	fields map[string]interface{}
}

// CaddyModule returns the Caddy module information.
func (LogMetrics) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "caddy.logging.encoders.prometheus",
		New: func() caddy.Module {
			m := new(LogMetrics)
			fmt.Printf("New: %v\n", &m)
			return m
		},
	}
}

// type zapLogger struct {
// 	zl *zap.Logger
// }

// func (l *zapLogger) Println(v ...interface{}) {
// 	l.zl.Sugar().Error(v...)
// }

// Provision -
func (m *LogMetrics) Provision(ctx caddy.Context) error {
	fmt.Printf("Provision: %v\n", &m)
	if m.WrappedRaw == nil {
		return fmt.Errorf("missing \"wrap\" (must specify an underlying encoder)")
	}

	log := ctx.Logger(m)
	m.metricsHandler = promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
		ErrorLog:      &zapLogger{log},
	})

	// set up wrapped encoder (required)
	val, err := ctx.LoadModule(m, "WrappedRaw")
	if err != nil {
		return fmt.Errorf("loading fallback encoder module: %w", err)
	}
	m.wrapped = val.(zapcore.Encoder)

	return m.initMetrics(ctx)
}

// ServeHTTP - instrument the handler
// fulfils the caddyhttp.MiddlewareHandler interface
func (m LogMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) (err error) {
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
	_ caddy.Provisioner           = (*LogMetrics)(nil)
	_ caddyhttp.MiddlewareHandler = (*LogMetrics)(nil)

	_ zapcore.Encoder = (*LogMetrics)(nil)
	// _ zapcore.Core    = (*LogMetrics)(nil)
)

// // Enabled - always enabled!
// func (*LogMetrics) Enabled(_ zapcore.Level) bool {
// 	return true
// }

// // With -
// func (m *LogMetrics) With(f []zapcore.Field) zapcore.Core {
// 	// no-op for now, though it may make sense to curry some of these fields?
// 	// not sure how good an idea that'd be...
// 	return m
// }

// // Check - no-op
// func (m *LogMetrics) Check(e zapcore.Entry, c *zapcore.CheckedEntry) *zapcore.CheckedEntry { return c }

// // Write - here's where the magic happens!
// func (m *LogMetrics) Write(e zapcore.Entry, f []zapcore.Field) error {
// 	return nil
// }

// // Sync -
// func (m *LogMetrics) Sync() error {
// 	return nil
// }

// AddArray is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	m.setField(key, marshaler)

	return m.wrapped.AddArray(key, marshaler)
}

func (m *LogMetrics) setField(key string, value interface{}) {
	if m.fields == nil {
		m.fields = make(map[string]interface{})
	}
	fmt.Printf("setField(%s, %v)\n", key, value)
	m.fields[key] = value
}

// AddObject is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	m.setField(key, marshaler)

	return m.wrapped.AddObject(key, marshaler)
}

// AddBinary is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddBinary(key string, value []byte) {
	m.setField(key, value)
	m.wrapped.AddBinary(key, value)
}

// AddByteString is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddByteString(key string, value []byte) {
	m.setField(key, value)
	m.wrapped.AddByteString(key, value)
}

// AddBool is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddBool(key string, value bool) {
	m.setField(key, value)
	m.wrapped.AddBool(key, value)
}

// AddComplex128 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddComplex128(key string, value complex128) {
	m.setField(key, value)
	m.wrapped.AddComplex128(key, value)
}

// AddComplex64 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddComplex64(key string, value complex64) {
	m.setField(key, value)
	m.wrapped.AddComplex64(key, value)
}

// AddDuration is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddDuration(key string, value time.Duration) {
	m.setField(key, value)
	m.wrapped.AddDuration(key, value)
}

// AddFloat64 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddFloat64(key string, value float64) {
	m.setField(key, value)
	m.wrapped.AddFloat64(key, value)
}

// AddFloat32 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddFloat32(key string, value float32) {
	m.setField(key, value)
	m.wrapped.AddFloat32(key, value)
}

// AddInt is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddInt(key string, value int) {
	m.setField(key, value)
	m.wrapped.AddInt(key, value)
}

// AddInt64 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddInt64(key string, value int64) {
	m.setField(key, value)
	m.wrapped.AddInt64(key, value)
}

// AddInt32 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddInt32(key string, value int32) {
	m.setField(key, value)
	m.wrapped.AddInt32(key, value)
}

// AddInt16 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddInt16(key string, value int16) {
	m.setField(key, value)
	m.wrapped.AddInt16(key, value)
}

// AddInt8 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddInt8(key string, value int8) {
	m.setField(key, value)
	m.wrapped.AddInt8(key, value)
}

// AddString is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddString(key, value string) {
	m.setField(key, value)

	m.wrapped.AddString(key, value)
}

// AddTime is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddTime(key string, value time.Time) {
	m.setField(key, value)
	m.wrapped.AddTime(key, value)
}

// AddUint is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddUint(key string, value uint) {
	m.setField(key, value)
	m.wrapped.AddUint(key, value)
}

// AddUint64 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddUint64(key string, value uint64) {
	m.setField(key, value)
	m.wrapped.AddUint64(key, value)
}

// AddUint32 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddUint32(key string, value uint32) {
	m.setField(key, value)
	m.wrapped.AddUint32(key, value)
}

// AddUint16 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddUint16(key string, value uint16) {
	m.setField(key, value)
	m.wrapped.AddUint16(key, value)
}

// AddUint8 is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddUint8(key string, value uint8) {
	m.setField(key, value)
	m.wrapped.AddUint8(key, value)
}

// AddUintptr is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddUintptr(key string, value uintptr) {
	m.setField(key, value)
	m.wrapped.AddUintptr(key, value)
}

// AddReflected is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) AddReflected(key string, value interface{}) error {
	m.setField(key, value)
	return m.wrapped.AddReflected(key, value)
}

// OpenNamespace is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) OpenNamespace(key string) {
	m.wrapped.OpenNamespace(key)
}

// Clone is part of the zapcore.ObjectEncoder interface.
func (m *LogMetrics) Clone() zapcore.Encoder {
	f := make(map[string]interface{})
	for k, v := range m.fields {
		f[k] = v
	}
	n := &LogMetrics{
		wrapped: m.wrapped.Clone(),
		fields:  f,
	}
	return n
}

// var bufferpool = buffer.NewPool()

// EncodeEntry partially implements the zapcore.Encoder interface.
func (m *LogMetrics) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// make sure these are cleared for the next encoder, I guess?
	defer func() { m.fields = nil }()

	fmt.Printf("(I am %v) got log for %s\n", &m, entry.LoggerName)
	fmt.Printf("I have fields: %#v\n\n", m.fields)
	logCount.WithLabelValues(entry.LoggerName).Inc()
	logFields.WithLabelValues(entry.LoggerName).Observe(float64(len(fields)))
	logAddedFields.WithLabelValues(entry.LoggerName).Observe(float64(len(m.fields)))

	if strings.HasPrefix(entry.LoggerName, "http.log.access.") {
		labels := prometheus.Labels{"code": "", "method": ""}
		var dur time.Duration
		var reqSize int64
		var resSize int64
		if r, ok := m.fields["request"]; ok {
			if lr, ok := r.(caddyhttp.LoggableHTTPRequest); ok {
				labels["method"] = lr.Method
				if lr.ContentLength > 0 {
					reqSize = lr.ContentLength
				}
			}
		}
		for _, field := range fields {
			switch field.Key {
			case "latency", "duration":
				dur = time.Duration(field.Integer)
			case "size":
				resSize = field.Integer
			case "status":
				labels["code"] = strconv.FormatInt(field.Integer, 10)
			case "method":
				labels["method"] = field.String
			}
		}

		requestDuration.With(labels).Observe(dur.Seconds())
		requestSize.With(labels).Observe(float64(reqSize))
		responseSize.With(labels).Observe(float64(resSize))
	}

	return m.wrapped.EncodeEntry(entry, fields)
}

/////////////

func (m *LogMetrics) initMetrics(ctx caddy.Context) error {
	log := ctx.Logger(m)

	m.registerMetrics("caddy", "http")
	if m.path == "" {
		m.path = defaultPath
	}
	if m.Addr == "" {
		m.Addr = defaultAddr
	}

	if !m.useCaddyAddr {
		mux := http.NewServeMux()
		mux.Handle(m.path, m.metricsHandler)

		srv := &http.Server{Handler: mux}
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

func (m *LogMetrics) registerMetrics(namespace, subsystem string) {
	if m.latencyBuckets == nil {
		m.latencyBuckets = append(prometheus.DefBuckets, 15, 20, 30, 60, 120, 180, 240, 480, 960)
	}
	if m.sizeBuckets == nil {
		m.sizeBuckets = []float64{0, 500, 1000, 2000, 3000, 4000, 5000, 10000, 20000, 30000, 50000, 1e5, 5e5, 1e6, 2e6, 3e6, 4e6, 5e6, 10e6}
	}

	// TODO: add "handler" and probably others
	httpLabels := []string{"code", "method"}

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

	requestSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_size_bytes",
		Help:      "Size of the request body in bytes",
		Buckets:   m.sizeBuckets,
	}, httpLabels)

	responseSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "response_size_bytes",
		Help:      "Size of the response in bytes",
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

	logCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: "log",
		Name:      "entry_count_total",
		Help:      "Counter of log messages",
	}, []string{"logger"})

	logFields = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "log",
		Name:      "field_counts",
		Help:      "histogram of log field counts",
		Buckets:   []float64{0, 1, 2, 3, 5, 8, 13, 21},
	}, []string{"logger"})

	logAddedFields = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: "log",
		Name:      "addedfield_counts",
		Help:      "histogram of log added field counts",
		Buckets:   []float64{0, 1, 2, 3, 5, 8, 13, 21},
	}, []string{"logger"})
}
