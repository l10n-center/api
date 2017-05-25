package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/pressly/chi"
	"github.com/pressly/chi/docgen"
	"github.com/pressly/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *server) router() chi.Router {
	r := chi.NewRouter()

	r.Use(tracingMiddleware)
	r.Use(s.authMiddleware())
	r.Use(loggerMiddleware)

	r.Route("/auth", func(r chi.Router) {
		r.Get("/", s.authCheck())
		r.Post("/init", s.authInit())
		r.Post("/login", s.authLogin())
		// r.Post("/forget", s.authForget())
		// r.Post("/reset/:token", s.authReset())
	})

	return r
}

// NewRouter api router
func NewRouter(db *sql.DB, secret []byte) chi.Router {
	s := newServer(db, secret)

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/debug", middleware.Profiler())
	r.Mount("/metrics", promhttp.Handler())
	r.Get("/doc", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(docgen.JSONRoutesDoc(s.router())))
	})

	r.Mount("/", s.router())

	return r
}
