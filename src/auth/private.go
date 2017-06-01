package auth

import (
	"net/http"

	"github.com/l10n-center/api/src/model"
	"github.com/l10n-center/api/src/tracing"
)

// Private check claims and given role to access control
func Private(role model.Role) func(http.Handler) http.Handler {
	// Private check claims and given role to access control
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			l := tracing.Logger(ctx)
			c, ok := ClaimsFromContext(ctx)

			if !ok {
				l.Warn("unauthorized access")
				w.WriteHeader(http.StatusUnauthorized)

				return
			}

			if role > 0 && c.Role&role == 0 {
				l.Warn("forbidden access")
				w.WriteHeader(http.StatusForbidden)

				return
			}
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
