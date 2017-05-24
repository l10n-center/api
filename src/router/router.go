package router

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/chi"
	sq "gopkg.in/Masterminds/squirrel.v1"
)

type server struct {
	db     *sqlx.DB
	qb     sq.StatementBuilderType
	secret []byte
}

// New api router
func New(db *sql.DB, secret []byte) chi.Router {
	qb := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		RunWith(db)

	dbx := sqlx.NewDb(db, "postgres")
	s := &server{
		qb:     qb,
		db:     dbx,
		secret: secret,
	}

	r := chi.NewRouter()
	r.Use(s.authMiddleware)
	r.Route("/auth", s.authRoute)

	return r
}
