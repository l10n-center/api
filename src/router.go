package api

import (
	"net/http"
	"time"

	"github.com/l10n-center/api/src/auth"
	"github.com/l10n-center/api/src/config"
	mw "github.com/l10n-center/api/src/middleware"
	"github.com/l10n-center/api/src/tracing"

	"github.com/pressly/chi"
	"github.com/pressly/chi/docgen"
	"github.com/pressly/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Store is a combined interface to model store
type Store interface {
	auth.Store
}

func router(cfg *config.Config, store Store) chi.Router {
	r := chi.NewRouter()

	r.Use(tracing.WithSpan)
	r.Use(mw.Boundary)

	r.Route("/auth", auth.Router(cfg, store))

	return r
}

// NewRouter api router
func NewRouter(cfg *config.Config, store Store) chi.Router {
	api := router(cfg, store)

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/debug", middleware.Profiler())
	r.Mount("/metrics", promhttp.Handler())
	r.Get("/doc", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(docgen.JSONRoutesDoc(api))); err != nil {
			zap.L().Error(err.Error())
		}
	})

	r.Mount("/", api)

	return r
}
