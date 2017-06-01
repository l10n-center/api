package controller

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/l10n-center/api/src/model"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type authStoreMock struct {
	mock.Mock
	users []*model.User
}

func (asm *authStoreMock) GetUserCount(_ context.Context) (int, error) {
	return len(asm.users), nil
}

func (asm *authStoreMock) GetUserByID(_ context.Context, id int32) (*model.User, error) {
	if id > 0 && id <= int32(len(asm.users)+1) {
		u := asm.users[id-1]
		u.ID = id

		return u, nil
	}

	return nil, sql.ErrNoRows
}

func (asm *authStoreMock) GetUserByEmail(_ context.Context, email string) (*model.User, error) {
	for i, u := range asm.users {
		if u.Email == email {
			u.ID = int32(i + 1)

			return u, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (asm *authStoreMock) CreateUser(_ context.Context, u *model.User) error {
	for _, su := range asm.users {
		if su.Email == u.Email {

			return errors.New("email collision")
		}
	}
	asm.users = append(asm.users, u)

	return nil
}

func TestAuthCheck(t *testing.T) {
	cases := []struct {
		name  string
		users []*model.User
		code  int
	}{
		{
			name:  "No users",
			users: []*model.User{},
			code:  http.StatusNotFound,
		},
		{
			name: "Login required",
			users: []*model.User{
				{
					Email:    "test@email.com",
					Passhash: []byte("$2a$04$jBtJJZgQXKf8vXtdBBktl.k9hJYWd6tPDSF4Tuz5SN7cvr9tZ8nu."),
					Role:     model.RoleAdmin,
				},
			},
			code: http.StatusUnauthorized,
		},
	}
	cfg := &Config{
		Secret: []byte{},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := &authStoreMock{users: c.users}
			handler := authCheck(cfg, store)
			req := httptest.NewRequest("GET", "/auth", &bytes.Buffer{})
			res := httptest.NewRecorder()
			handler(res, req)
			require.Equal(t, c.code, res.Code, res.Body.String())
		})
	}
}
