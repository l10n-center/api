package server

import (
	"database/sql"
	"time"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter api router
func NewRouter(db *sql.DB, secret []byte) chi.Router {
	s := newServer(db, secret)

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(tracingMiddleware)
	r.Use(s.authMiddleware)
	r.Use(loggerMiddleware)

	r.Mount("/debug", middleware.Profiler())
	r.Mount("/metrics", promhttp.Handler())

	r.Route("/auth", func(r chi.Router) {
		r.Get("/", s.authCheck)
		r.Post("/init", s.authInit)
		r.Post("/login", s.authLogin)
		// r.Post("/forget", s.authForget)
		// r.Post("/reset/:token", s.authReset)
	})

	return r
}
