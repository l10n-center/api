package auth

import (
	"context"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/l10n-center/api/src/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type claimsCtxKey struct{}

// Claims of authorization token
type Claims struct {
	jwt.StandardClaims
	UserID  bson.ObjectId `json:"userId"`
	Email   string        `json:"email"`
	IsAdmin bool          `json:"isAdmin"`
}

// ContextWithClaims store claims in context
func ContextWithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsCtxKey{}, c)
}

// ClaimsFromContext try to extract authorization claims from context
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsCtxKey{}).(*Claims)

	return c, ok
}

// CreateToken from user and sign it
func CreateToken(ctx context.Context, secret string, u *model.User) (string, error) {
	c := &Claims{
		UserID:  u.ID,
		Email:   u.Email,
		IsAdmin: u.IsAdmin,
	}
	c.ExpiresAt = time.Now().AddDate(0, 0, 14).Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	st, err := t.SignedString([]byte(secret))
	if err != nil {

		return "", errors.WithStack(err)
	}

	return st, nil
}
