package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/l10n-center/api/src/auth"
	"github.com/l10n-center/api/src/model"
	"github.com/l10n-center/api/src/tracing"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type authStore interface {
	GetUserCount(context.Context) (int, error)
	GetUserByID(context.Context, int32) (*model.User, error)
	GetUserByEmail(context.Context, string) (*model.User, error)
	CreateUser(context.Context, *model.User) error
}

func authCheck(cfg *Config, store authStore) http.HandlerFunc {
	// check login availability and update jwt if present valid
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := tracing.Logger(ctx)

		je := json.NewEncoder(w)

		userCount, err := store.GetUserCount(ctx)
		if err != nil {
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
		c, ok := auth.ClaimsFromContext(ctx)
		if !ok {
			l.Info("login required")
			w.WriteHeader(http.StatusUnauthorized)
			je.Encode("login required")
			return
		}
		u, err := store.GetUserByID(ctx, c.UserID)
		if err != nil {
			if errors.Cause(err) == sql.ErrNoRows {
				w.WriteHeader(http.StatusUnauthorized)
				je.Encode("login required")
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			je.Encode("login required")
			return

		}
		st, err := auth.CreateToken(ctx, cfg.Secret, u)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			je.Encode([]string{"something went wrong"})
			return
		}
		w.WriteHeader(http.StatusOK)
		je.Encode(st)
	}
	return http.HandlerFunc(fn)
}

func authLogin(cfg *Config, store authStore) http.HandlerFunc {
	// login with email/password credentials
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := tracing.Logger(ctx)

		je := json.NewEncoder(w)
		jd := json.NewDecoder(r.Body)

		rd := &struct {
			Email    string `json:"email" valid:"email,required"`
			Password string `json:"password" valid:"required"`
		}{}
		if err := jd.Decode(&rd); err != nil {
			l.Info(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			je.Encode([]string{err.Error()})
			return
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
		u, err := store.GetUserByEmail(ctx, rd.Email)
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
			je.Encode([]string{"invalid email or password"})
			return
		} else if err != nil {
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
		l.Sugar().Infof("user <%s> is logined", rd.Email)
		st, err := auth.CreateToken(ctx, cfg.Secret, u)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			je.Encode([]string{"something went wrong"})
			return
		}
		w.WriteHeader(http.StatusOK)
		je.Encode(st)
	}
	return http.HandlerFunc(fn)
}

func authInit(_ *Config, store authStore) http.HandlerFunc {
	// create admin user
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := tracing.Logger(ctx)

		w.Header().Set("Content-Type", "application/json")
		je := json.NewEncoder(w)
		jd := json.NewDecoder(r.Body)

		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			je.Encode("accept only application/json")
			return
		}
		rd := &struct {
			Email    string `json:"email" valid:"email,required"`
			Password string `json:"password" valid:"required"`
		}{}
		if err := jd.Decode(&rd); err != nil {
			l.Info(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			je.Encode([]string{err.Error()})
			return
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

		passhash, err := bcrypt.GenerateFromPassword([]byte(rd.Password), bcrypt.DefaultCost)
		if err != nil {
			l.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			je.Encode("something went wrong")
			return
		}

		userCount, err := store.GetUserCount(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			je.Encode([]string{"something went wrong"})
			return
		}
		if userCount > 0 {
			l.Warn("admin already exists")
			w.WriteHeader(http.StatusForbidden)
			je.Encode("admin already exists")
			return
		}

		err = store.CreateUser(ctx, &model.User{
			Email:    rd.Email,
			Passhash: passhash,
			Role:     model.RoleAdmin,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			je.Encode("something went wrong")
			return
		}
		l.Sugar().Infof("admin created with email <%s>", rd.Email)
		w.WriteHeader(http.StatusCreated)
		je.Encode("admin created")
	}
	return http.HandlerFunc(fn)
}
