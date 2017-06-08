package middleware

import (
	"net/http"

	"github.com/l10n-center/api/src/tracing"
)

// JSONOnly set response content type to application/json
// and if request method POST, PUT or PATCH check request
// content type and return error if it not json
func JSONOnly(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := tracing.Logger(ctx)

		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			if r.Header.Get("Content-Type") != "application/json" {
				l.Debug("not json request")
				http.Error(w, "Content-Type not allowed. Use application/json", http.StatusBadRequest)

				return
			}
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
