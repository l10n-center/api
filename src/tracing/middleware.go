package tracing

import (
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pressly/chi/middleware"
	"github.com/uber/jaeger-client-go"
)

// Middleware to add tracing span to context and it's id in response header
func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sp := opentracing.StartSpan(r.URL.Path)
		id := sp.Context().(jaeger.SpanContext).TraceID().String()
		ctx = opentracing.ContextWithSpan(ctx, sp)
		ctx = ContextWithTraceID(ctx, id)
		w.Header().Set("Trace-ID", id)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			if ww.Status() >= 400 {
				ext.Error.Set(sp, true)
			}
			sp.Finish()
		}()

		next.ServeHTTP(ww, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
