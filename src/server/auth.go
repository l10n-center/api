package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/l10n-center/api/src/model"

	"github.com/asaskevich/govalidator"
	"github.com/dgrijalva/jwt-go"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type tokenCtxKey struct{}

type claims struct {
	jwt.StandardClaims
	Email string     `json:"email"`
	Role  model.Role `json:"role"`
}

func (s *server) authMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) > 7 && strings.ToUpper(authHeader[:7]) == "BEARER " {
			l := tracedLogger(ctx)
			sp := opentracing.SpanFromContext(ctx)
			tokenString := authHeader[7:]
			token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
				if t.Method != jwt.SigningMethodHS256 {

					return nil, fmt.Errorf("unexpected signing method %q", t.Method)
				}

				return s.secret, nil
			})
			if err != nil {
				l.Error(err.Error())
			} else if token.Valid {
				c := token.Claims.(claims)
				sp.SetTag("claims", c)
				ctx = context.WithValue(r.Context(), tokenCtxKey{}, token)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

func (s *server) createToken(ctx context.Context, email string, role *model.Role) (string, error) {
	c := claims{
		Email: email,
	}
	c.ExpiresAt = time.Now().AddDate(0, 0, 14).Unix()
	if role == nil {
		err := traceDB(ctx, "getRole", func() error {
			return s.qb.
				Select("role").
				From("public.user").
				Where("email = ?", email).
				QueryRow().
				Scan(&c.Role)
		})
		if err != nil {
			return "", errors.Wrap(err, "role query")
		}
	} else {
		c.Role = *role
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	st, err := t.SignedString(s.secret)
	if err != nil {
		return "", errors.Wrap(err, "signing token")
	}
	return st, nil
}

func (s *server) authCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := tracedLogger(ctx)

	w.Header().Set("Content-Type", "application/json")
	je := json.NewEncoder(w)

	var userCount int

	err := traceDB(ctx, "countUsers", func() error {
		return s.qb.
			Select("count(*)").
			From("public.user").
			Limit(1).
			QueryRow().
			Scan(&userCount)
	})
	if err != nil {
		l.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		je.Encode([]string{"something went wrong"})
		return
	}
	if userCount == 0 {
		l.Warn("users not found")
		w.WriteHeader(http.StatusNotFound)
		je.Encode("users not found")
		return
	}
	t, ok := r.Context().Value(tokenCtxKey{}).(*jwt.Token)
	if !ok {
		l.Info("login required")
		w.WriteHeader(http.StatusUnauthorized)
		je.Encode("login required")
		return
	}
	c := t.Claims.(*claims)
	st, err := s.createToken(ctx, c.Email, nil)
	if err != nil {
		l.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		je.Encode([]string{"something went wrong"})
		return
	}
	w.WriteHeader(http.StatusOK)
	je.Encode(st)
}

func (s *server) authLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := tracedLogger(ctx)

	w.Header().Set("Content-Type", "application/json")
	je := json.NewEncoder(w)

	rd := &struct {
		Email    string `json:"email" valid:"email,required"`
		Password string `json:"password" valid:"required"`
	}{}
	if r.Header.Get("Content-Type") == "application/json" {
		d := json.NewDecoder(r.Body)

		if err := d.Decode(&rd); err != nil {
			l.Info(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			je.Encode([]string{err.Error()})
			return
		}
	} else {
		rd.Email = r.FormValue("email")
		rd.Password = r.FormValue("password")
	}
	if ok, err := govalidator.ValidateStruct(rd); !ok {
		l.Info(err.Error())
		res := make([]string, len(err.(govalidator.Errors)))
		for i, e := range err.(govalidator.Errors) {
			res[i] = e.Error()
		}
		w.WriteHeader(http.StatusBadRequest)
		je.Encode(res)
		return
	}
	u := &model.User{}
	query, args, err := s.qb.
		Select("role, passhash").
		From("public.user").
		Where("email = ?", rd.Email).
		ToSql()
	if err != nil {
		l.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		je.Encode([]string{"something went wrong"})
		return
	}
	err = traceDB(ctx, "getUser", func() error {
		return s.db.Get(u, query, args...)
	})
	if err == sql.ErrNoRows {
		l.Sugar().Infof("user '%s' is not found", rd.Email)
		w.WriteHeader(http.StatusBadRequest)
		je.Encode([]string{"invalid email or password"})
		return
	} else if err != nil {
		l.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		je.Encode([]string{"something went wrong"})
		return
	}
	if err := bcrypt.CompareHashAndPassword(u.Passhash, []byte(rd.Password)); err != nil {
		l.Info("invalid password")
		w.WriteHeader(http.StatusBadRequest)
		je.Encode([]string{"invalid email or password"})
		return
	}
	l.Sugar().Infof("user '%s' is logined", rd.Email)
	st, err := s.createToken(ctx, rd.Email, &u.Role)
	if err != nil {
		l.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		je.Encode([]string{"something went wrong"})
		return
	}
	http.Error(w, st, http.StatusOK)

}

func (s *server) authInit(w http.ResponseWriter, r *http.Request) {}
