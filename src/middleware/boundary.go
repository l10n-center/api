package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/l10n-center/api/src/tracing"
	"github.com/pressly/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	responseStatusM = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "l10n_center",
			Subsystem: "api",
			Name:      "response_per_status",
			Help:      "Count of responses per status hundred code",
		},
		[]string{"status"},
	)
	responseDurationM = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "l10n_center",
			Subsystem: "api",
			Name:      "response_duration",
			Help:      "Duration of response",
			Buckets:   prometheus.LinearBuckets(0, 10, 10),
		},
	)
)

func init() {
	prometheus.MustRegister(responseStatusM)
	prometheus.MustRegister(responseDurationM)
}

// Boundary log start and end of request and store metrics of request
//
// Require tracing.WithSpan
func Boundary(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := tracing.Logger(ctx)

		l.Info(
			"request start",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("clientIP", r.RemoteAddr),
			zap.String("userAgent", r.UserAgent()),
		)

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		ts := time.Now()
		defer func() {
			te := time.Now()
			l.Info(
				"request end",
				zap.Int("status", ww.Status()),
				zap.Int("responseLength", ww.BytesWritten()),
				zap.Duration("duration", te.Sub(ts)),
			)
			responseStatusM.WithLabelValues(strconv.Itoa(ww.Status() / 100 * 100)).Add(1)
			responseDurationM.Observe(float64(te.Sub(ts)) / float64(time.Millisecond))
		}()

		next.ServeHTTP(ww, r)
	}

	return http.HandlerFunc(fn)
}
