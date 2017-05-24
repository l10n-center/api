package router

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/l10n-center/api/src/model"
	"github.com/pressly/chi"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type ctxToken struct{}

type claims struct {
	jwt.StandardClaims
	Email string     `json:"email"`
	Role  model.Role `json:"role"`
}

func (s *server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) > 7 && strings.ToUpper(authHeader[:7]) == "BEARER " {
			tokenString := authHeader[7:]
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if t.Method != jwt.SigningMethodHS256 {
					return nil, fmt.Errorf("unexpected signing method %q", t.Method)
				}

				return s.secret, nil
			})
			if err != nil {
				log.Printf("ERROR: %s", err)
			} else if token.Valid {
				ctx = context.WithValue(r.Context(), ctxToken{}, token)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *server) createToken(email string) (string, error) {
	c := claims{
		Email: email,
	}
	c.ExpiresAt = time.Now().AddDate(0, 0, 14).Unix()
	err := s.qb.
		Select("role").
		From("public.user").
		Where("email = ?", email).
		QueryRow().
		Scan(&c.Role)
	if err != nil {
		return "", err
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	return t.SignedString(s.secret)
}

func (s *server) authRoute(r chi.Router) {
	r.Get("/", s.authCheck)
	r.Post("/init", s.authInit)
	r.Post("/login", s.authLogin)
	// r.Post("/forget", s.authForget)
	// r.Post("/reset/:token", s.authReset)
}

func (s *server) authCheck(w http.ResponseWriter, r *http.Request) {
	var userCount int

	err := s.qb.
		Select("count(*)").
		From("public.user").
		Limit(1).
		QueryRow().
		Scan(&userCount)

	if err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	if userCount == 0 {
		http.Error(w, "users not found", http.StatusNotFound)
		return
	}
	t, ok := r.Context().Value(ctxToken{}).(*jwt.Token)
	if !ok {
		http.Error(w, "login required", http.StatusUnauthorized)
		return
	}
	c, ok := t.Claims.(*claims)
	if !ok {
		http.Error(w, "login required", http.StatusUnauthorized)
		return
	}
	st, err := s.createToken(c.Email)
	if err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	http.Error(w, st, http.StatusOK)
}

func (s *server) authLogin(w http.ResponseWriter, r *http.Request) {
	rd := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if r.Header.Get("Content-Type") == "application/json" {
		d := json.NewDecoder(r.Body)

		if err := d.Decode(&rd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		rd.Email = r.FormValue("email")
		rd.Password = r.FormValue("password")
	}
	if len(rd.Email) == 0 {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if len(rd.Password) == 0 {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}
	u := &model.User{}
	query, args, err := s.qb.
		Select("passhash").
		From("public.user").
		Where("email = $1", rd.Email).
		ToSql()
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	if err := s.db.Get(u, query, args...); err == sql.ErrNoRows {
		http.Error(w, "invalid email or password", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword(u.Passhash, []byte(rd.Password)); err != nil {
		http.Error(w, "invalid email or password", http.StatusBadRequest)
		return
	}
	log.Printf("INFO: %s is logined", rd.Email)
}

func (s *server) authInit(w http.ResponseWriter, r *http.Request) {}
