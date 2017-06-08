package auth

import (
	"context"

	"github.com/l10n-center/api/src/model"

	"gopkg.in/mgo.v2/bson"
)

//go:generate mockgen -source=store.go -destination=store_mock_test.go -package=auth_test

// Store is a interface of store required in package auth
type Store interface {
	GetUserCount(context.Context) (int, error)
	GetUserByID(context.Context, bson.ObjectId) (*model.User, error)
	GetUserByEmail(context.Context, string) (*model.User, error)
	CreateUser(context.Context, *model.User) error
}
