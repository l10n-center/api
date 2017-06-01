package controller

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/l10n-center/api/src/auth"
	mw "github.com/l10n-center/api/src/middleware"
	"github.com/l10n-center/api/src/model"
	"github.com/l10n-center/api/src/tracing"

	"github.com/pressly/chi"
	"github.com/pressly/chi/docgen"
	"github.com/pressly/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config of application
type Config struct {
	Secret []byte
}

func apiRouter(cfg *Config, store *model.Store) chi.Router {
	r := chi.NewRouter()

	r.Use(tracing.Middleware)
	r.Use(mw.Boundary)
	r.Use(auth.Middleware(cfg.Secret))

	r.Route("/auth", func(r chi.Router) {
		r.Use(mw.JSONOnly)

		r.Get("/", authCheck(cfg, store))
		r.Post("/init", authInit(cfg, store))
		r.Post("/login", authLogin(cfg, store))
		// r.Post("/forget", s.authForget())
		// r.Post("/reset/:token", s.authReset())
	})

	return r
}

// NewRouter api router
func NewRouter(cfg *Config, db *sql.DB) chi.Router {
	store := model.NewStore(db)
	api := apiRouter(cfg, store)
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/debug", middleware.Profiler())
	r.Mount("/metrics", promhttp.Handler())
	r.Get("/doc", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(docgen.JSONRoutesDoc(api)))
	})

	r.Mount("/", api)

	return r
}
