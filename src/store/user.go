package store

import (
	"context"

	"github.com/l10n-center/api/src/errs"
	"github.com/l10n-center/api/src/model"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const userCollection = "user"

func (s *Store) initUser() error {
	err := s.mongo.DB("").C(userCollection).Create(&mgo.CollectionInfo{
		Validator: bson.M{"username": bson.M{
			"$type":   "string",
			"$exists": true,
		}},
	})

	if err == nil {
		err = s.mongo.DB("").C(userCollection).EnsureIndex(mgo.Index{
			Key:    []string{"email"},
			Unique: true,
		})
	}

	return errors.WithStack(err)
}

// GetUserCount return count of users
func (s *Store) GetUserCount(ctx context.Context) (int, error) {
	sp, _ := opentracing.StartSpanFromContext(ctx, "db:GetUserCount")

	defer sp.Finish()

	m := s.mongo.Clone()

	defer m.Close()

	n, err := m.DB("").C(userCollection).Count()

	return n, errors.WithStack(err)
}

// GetUserByID search user collection by id
func (s *Store) GetUserByID(ctx context.Context, id bson.ObjectId) (*model.User, error) {
	sp, _ := opentracing.StartSpanFromContext(ctx, "db:GetUserByID")

	defer sp.Finish()

	m := s.mongo.Clone()

	defer m.Close()

	u := &model.User{}

	err := m.DB("").C(userCollection).FindId(id).One(u)

	if err == mgo.ErrNotFound {
		err = errs.ModelNotFound
	}

	return u, errors.WithStack(err)
}

// GetUserByEmail search user by email
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	sp, _ := opentracing.StartSpanFromContext(ctx, "db:GetUserByEmail")

	defer sp.Finish()

	m := s.mongo.Clone()

	defer m.Close()

	u := &model.User{}

	err := m.DB("").C(userCollection).Find(bson.M{"email": email}).One(u)

	if err == mgo.ErrNotFound {
		err = errs.ModelNotFound
	}

	return u, errors.WithStack(err)
}

// CreateUser insert new user
func (s *Store) CreateUser(ctx context.Context, u *model.User) error {
	sp, _ := opentracing.StartSpanFromContext(ctx, "db:CreateUser")

	defer sp.Finish()

	m := s.mongo.Clone()

	defer m.Close()

	return errors.WithStack(m.DB("").C(userCollection).Insert(u))
}
