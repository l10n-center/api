package server

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	sq "gopkg.in/Masterminds/squirrel.v1"
)

type server struct {
	db     *sqlx.DB
	qb     sq.StatementBuilderType
	secret []byte
}

func newServer(db *sql.DB, secret []byte) *server {
	qb := sq.
		StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		RunWith(db)

	dbx := sqlx.NewDb(db, "postgres")

	return &server{
		qb:     qb,
		db:     dbx,
		secret: secret,
	}
}
