package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/l10n-center/api/src/config"
	"github.com/l10n-center/api/src/errs"
	"github.com/l10n-center/api/src/tracing"

	"github.com/dgrijalva/jwt-go"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// WithClaims middleware check Authorization header and try to parse token
//
// Require tracing.WithSpan
func WithClaims(cfg *config.Config, required, onlyAdmin bool) func(http.Handler) http.Handler {
	// WithClaims check Authorization header and try to parse token
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var c *Claims
			ctx := r.Context()
			l := tracing.Logger(ctx)
			authHeader := r.Header.Get("Authorization")
			if len(authHeader) > 7 && strings.ToUpper(authHeader[:7]) == "BEARER " {
				tokenString := authHeader[7:]
				token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
					if t.Method != jwt.SigningMethodHS256 {

						return nil, errors.Errorf("unexpected signing method %q", t.Method)
					}

					return []byte(cfg.Secret), nil
				})
				if err == nil && token.Valid {
					c = token.Claims.(*Claims)
					if onlyAdmin {
						l.Debug("no admin")
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}
					ctx = context.WithValue(r.Context(), claimsCtxKey{}, c)
					sp := opentracing.SpanFromContext(ctx)
					sp = sp.SetTag("claims", c)
					ctx = opentracing.ContextWithSpan(ctx, sp)
				} else {
					l.Debug("bad token", zap.Error(err), errs.ZapStack(err))
					if required {
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}
				}
			} else if required {
				l.Debug("unauthorized: bad header")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
