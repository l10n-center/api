package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/l10n-center/api/src/tracing"

	"github.com/dgrijalva/jwt-go"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// Middleware to check Authorization header and try to parse token
func Middleware(secret []byte) func(http.Handler) http.Handler {
	// WithClaims check Authorization header and try to parse token
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var c *Claims
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) > 7 && strings.ToUpper(authHeader[:7]) == "BEARER " {
				l := tracing.Logger(ctx)
				sp := opentracing.SpanFromContext(ctx)
				tokenString := authHeader[7:]
				token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
					if t.Method != jwt.SigningMethodHS256 {

						return nil, errors.Errorf("unexpected signing method %q", t.Method)
					}

					return secret, nil
				})
				if err != nil {
					l.Error(err.Error())
				} else if token.Valid {
					c = token.Claims.(*Claims)
					sp.SetTag("claims", c)
					ctx = context.WithValue(r.Context(), claimsCtxKey{}, c)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
