package server

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pressly/chi/middleware"
)

func tracingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sp := opentracing.StartSpan(r.URL.Path)
		ctx = opentracing.ContextWithSpan(ctx, sp)
		sp.Tracer().Inject(
			sp.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(w.Header()),
		)
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

func traceDB(ctx context.Context, name string, fn func() error) error {
	sp, _ := opentracing.StartSpanFromContext(ctx, "db:"+name)

	defer sp.Finish()

	ext.SpanKind.Set(sp, "resources")
	ext.PeerService.Set(sp, "PostgreSQL")

	err := fn()
	if err != nil {
		ext.Error.Set(sp, true)
		sp.LogEventWithPayload("error", err)
	}

	return err
}
