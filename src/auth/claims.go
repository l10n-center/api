package auth

import (
	"context"

	"github.com/l10n-center/api/src/model"

	"github.com/dgrijalva/jwt-go"
)

type claimsCtxKey struct{}

// Claims of authorization token
type Claims struct {
	jwt.StandardClaims
	UserID int32      `json:"userId"`
	Email  string     `json:"email"`
	Role   model.Role `json:"role"`
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
