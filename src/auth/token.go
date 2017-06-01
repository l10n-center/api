package auth

import (
	"context"
	"time"

	"github.com/l10n-center/api/src/model"
	"github.com/l10n-center/api/src/tracing"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// CreateToken from user and sign it
func CreateToken(ctx context.Context, secret []byte, u *model.User) (string, error) {
	l := tracing.Logger(ctx)
	c := &Claims{
		UserID: u.ID,
		Email:  u.Email,
		Role:   u.Role,
	}
	c.ExpiresAt = time.Now().AddDate(0, 0, 14).Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	st, err := t.SignedString(secret)
	if err != nil {
		l.Error(err.Error())

		return "", errors.WithStack(err)
	}

	return st, nil
}
