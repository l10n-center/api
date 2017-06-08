package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/l10n-center/api/src/config"
	"github.com/l10n-center/api/src/errs"
	"github.com/l10n-center/api/src/middleware"
	"github.com/l10n-center/api/src/model"
	"github.com/l10n-center/api/src/tracing"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"github.com/pressly/chi"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Router return auth section router
func Router(cfg *config.Config, store Store) func(chi.Router) {

	return func(r chi.Router) {
		r.Use(WithClaims(cfg, false, false))
		r.Use(middleware.JSONOnly)

		r.Get("/", Root(cfg, store))
		r.Post("/init", Init(cfg, store))
		r.Post("/login", Login(cfg, store))
		// r.Post("/forgot")
		// r.Post("/reset")
	}
}

// Root check login availability and update jwt if present valid
func Root(cfg *config.Config, store Store) http.HandlerFunc {
	// Check login availability and update jwt if present valid
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := tracing.Logger(ctx)

		userCount, err := store.GetUserCount(ctx)
		if err != nil {
			l.Error(err.Error(), errs.ZapStack(err))
			http.Error(w, errs.InternalServerError, http.StatusInternalServerError)
			return
		}
		if userCount == 0 {
			l.Warn("users not found")
			http.Error(w, "users not found", http.StatusNotFound)
			return
		}
		c, ok := ClaimsFromContext(ctx)
		if !ok {
			l.Debug("no claims found")
			http.Error(w, "login required", http.StatusUnauthorized)
			return
		}
		u, err := store.GetUserByID(ctx, c.UserID)
		if err != nil {
			if errors.Cause(err) == errs.ModelNotFound {
				l.Debug("login required", zap.Error(err))
				http.Error(w, "login required", http.StatusUnauthorized)
				return
			}
			l.Error(err.Error(), errs.ZapStack(err))
			http.Error(w, errs.InternalServerError, http.StatusInternalServerError)
			return

		}
		st, err := CreateToken(ctx, cfg.Secret, u)
		if err != nil {
			l.Error(err.Error(), errs.ZapStack(err))
			http.Error(w, errs.InternalServerError, http.StatusInternalServerError)
			return
		}
		http.Error(w, st, http.StatusOK)
	}
	return http.HandlerFunc(fn)
}

// Login with email/password credentials
func Login(cfg *config.Config, store Store) http.HandlerFunc {
	// Login with email/password credentials
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := tracing.Logger(ctx)

		jd := json.NewDecoder(r.Body)

		rd := &struct {
			Email    string `json:"email" valid:"email,required"`
			Password string `json:"password" valid:"required"`
		}{}
		if err := errors.WithMessage(jd.Decode(&rd), "unmarshal"); err != nil {
			l.Debug(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if ok, err := govalidator.ValidateStruct(rd); !ok {
			l.Debug(err.Error())
			http.Error(w, errors.WithMessage(err, "validate").Error(), http.StatusBadRequest)
			return
		}
		u, err := store.GetUserByEmail(ctx, rd.Email)
		if errors.Cause(err) == errs.ModelNotFound {
			l.Debug(err.Error(), zap.Error(err))
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		} else if err != nil {
			l.Error(err.Error(), errs.ZapStack(err))
			http.Error(w, errs.InternalServerError, http.StatusInternalServerError)
			return
		}
		if err = bcrypt.CompareHashAndPassword(u.Passhash, []byte(rd.Password)); err != nil {
			l.Debug("invalid password", zap.Error(err))
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		st, err := CreateToken(ctx, cfg.Secret, u)
		if err != nil {
			l.Error(err.Error(), errs.ZapStack(err))
			http.Error(w, errs.InternalServerError, http.StatusInternalServerError)
			return
		}
		l.Sugar().Infof("user <%s> is logined", rd.Email)
		http.Error(w, st, http.StatusOK)
	}
	return http.HandlerFunc(fn)
}

// Init create admin user
func Init(_ *config.Config, store Store) http.HandlerFunc {
	// Create admin user
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := tracing.Logger(ctx)

		jd := json.NewDecoder(r.Body)

		rd := &struct {
			Email    string `json:"email" valid:"email,required"`
			Password string `json:"password" valid:"required"`
		}{}
		if err := errors.Wrap(jd.Decode(&rd), "unmarshal"); err != nil {
			l.Debug(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if ok, err := govalidator.ValidateStruct(rd); !ok {
			l.Debug(err.Error())
			http.Error(w, errors.Wrap(err, "validate").Error(), http.StatusBadRequest)
			return
		}

		passhash, err := bcrypt.GenerateFromPassword([]byte(rd.Password), bcrypt.DefaultCost)
		if err != nil {
			l.Error(err.Error())
			http.Error(w, errs.InternalServerError, http.StatusInternalServerError)
			return
		}

		userCount, err := store.GetUserCount(ctx)
		if err != nil {
			l.Error(err.Error(), errs.ZapStack(err))
			http.Error(w, errs.InternalServerError, http.StatusInternalServerError)
			return
		}
		if userCount > 0 {
			l.Warn("users already exists")
			http.Error(w, "users already exists", http.StatusForbidden)
			return
		}

		err = store.CreateUser(ctx, &model.User{
			Email:      rd.Email,
			Passhash:   passhash,
			IsAdmin:    true,
			Permission: model.CanEverything,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		})
		if err != nil {
			l.Error(err.Error(), errs.ZapStack(err))
			http.Error(w, errs.InternalServerError, http.StatusInternalServerError)
			return
		}
		l.Sugar().Infof("admin created with email <%s>", rd.Email)
		http.Error(w, "admin created", http.StatusCreated)
	}
	return http.HandlerFunc(fn)
}
