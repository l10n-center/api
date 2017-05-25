package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/l10n-center/api/src/server"

	"encoding/json"

	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	db := initDB("auth")

	defer db.Close()

	r := server.NewRouter(db, []byte{})

	t.Run("No user check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		res := w.Result()
		body := readBody(res)
		printUberTraceID(t, res)
		require.Equal(t, http.StatusNotFound, res.StatusCode, string(body))
	})

	t.Run("Create admin", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/init", bytes.NewBufferString(`{
			"email": "admin@mail.com",
			"password": "admin"
		}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		res := w.Result()
		body := readBody(res)
		printUberTraceID(t, res)
		require.Equal(t, http.StatusCreated, res.StatusCode, string(body))
	})

	t.Run("Recreate admin", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/init", bytes.NewBufferString(`{
			"email": "fakeAdmin@mail.com",
			"password": "admin"
		}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		res := w.Result()
		body := readBody(res)
		printUberTraceID(t, res)
		require.Equal(t, http.StatusForbidden, res.StatusCode, string(body))
	})

	t.Run("Login available", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		res := w.Result()
		body := readBody(res)
		printUberTraceID(t, res)
		require.Equal(t, http.StatusUnauthorized, res.StatusCode, string(body))
	})

	token := ""

	t.Run("Login", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString(`{
			"email": "admin@mail.com",
			"password": "admin"
		}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		res := w.Result()
		body := readBody(res)
		printUberTraceID(t, res)
		require.Equal(t, http.StatusOK, res.StatusCode, string(body))
		require.NoError(t, json.Unmarshal(body, &token))
	})

	t.Run("Already loggined", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth", nil)
		req.Header.Set("Authorization", "bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		res := w.Result()
		body := readBody(res)
		printUberTraceID(t, res)
		require.Equal(t, http.StatusOK, res.StatusCode, string(body))
	})
}
