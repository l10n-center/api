package middleware

import (
	"encoding/json"
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

		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			if r.Header.Get("Content-Type") != "application/json" {
				l.Warn("not json request")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode("Content-Type not allowed. Use application/json")

				return
			}
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
