package model

import (
	"context"
	"database/sql"

	"github.com/l10n-center/api/src/tracing"

	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gopkg.in/Masterminds/squirrel.v1"
)

// Store is an db wrapper to implement controller dependencies
type Store struct {
	*sqlx.DB
	qb squirrel.StatementBuilderType
}

// NewStore is a constructor
func NewStore(db *sql.DB) *Store {
	qb := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Dollar).
		RunWith(db)

	dbx := sqlx.NewDb(db, "postgres")

	return &Store{dbx, qb}
}

// GetUserCount return count of users in db
func (s *Store) GetUserCount(ctx context.Context) (int, error) {
	l := tracing.Logger(ctx)
	sp, ctx := opentracing.StartSpanFromContext(ctx, "db:UserCount")

	defer sp.Finish()

	query, args, err := s.qb.
		Select("count(*)").
		From("public.user").
		ToSql()
	if err != nil {
		l.Error(err.Error())

		return 0, errors.WithStack(err)
	}

	var c int
	if err := s.QueryRowContext(ctx, query, args...).Scan(&c); err != nil {
		l.Error(err.Error())

		return 0, errors.WithStack(err)
	}

	return c, nil
}

// GetUserByID return user by id
func (s *Store) GetUserByID(ctx context.Context, id int32) (*User, error) {
	l := tracing.Logger(ctx)
	sp, ctx := opentracing.StartSpanFromContext(ctx, "db:UserByID")

	defer sp.Finish()

	query, args, err := s.qb.
		Select("*").
		From("public.user").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		l.Error(err.Error())

		return nil, errors.WithStack(err)
	}

	u := &User{}
	if err := s.GetContext(ctx, u, query, args...); err != nil {
		l.Error(err.Error())

		return nil, errors.WithStack(err)
	}

	return u, nil
}

// GetUserByEmail return user by email
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	l := tracing.Logger(ctx)
	sp, ctx := opentracing.StartSpanFromContext(ctx, "db:UserByEmail")

	defer sp.Finish()

	query, args, err := s.qb.
		Select("*").
		From("public.user").
		Where("email = ?", email).
		ToSql()
	if err != nil {
		l.Error(err.Error())

		return nil, errors.WithStack(err)
	}

	u := &User{}
	if err := s.GetContext(ctx, u, query, args...); err != nil {
		l.Error(err.Error())

		return nil, errors.WithStack(err)
	}

	return u, nil
}

// CreateUser in db
func (s *Store) CreateUser(ctx context.Context, u *User) error {
	l := tracing.Logger(ctx)
	sp, ctx := opentracing.StartSpanFromContext(ctx, "db:UserByEmail")

	defer sp.Finish()

	query, args, err := s.qb.
		Insert("public.user").
		Columns("email", "passhash", "role").
		Values(u.Email, u.Passhash, u.Role).
		ToSql()
	if err != nil {
		l.Error(err.Error())

		return errors.WithStack(err)
	}

	if _, err := s.ExecContext(ctx, query, args...); err != nil {
		l.Error(err.Error())

		return errors.WithStack(err)
	}

	return nil
}
