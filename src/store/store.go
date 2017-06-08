package store

import (
	"github.com/l10n-center/api/src/config"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2"
)

// Store of data in mongo
type Store struct {
	mongo *mgo.Session
}

// New Store
func New(cfg *config.Config) (*Store, error) {
	mongo, err := mgo.Dial(cfg.MongoHost + "/" + cfg.MongoDB)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	mongo.SetMode(mgo.Monotonic, true)

	s := &Store{mongo}

	if err := s.initUser(); err != nil {

		return nil, errors.WithStack(err)
	}

	return s, nil
}

// Close mongo session
func (s *Store) Close() {
	s.mongo.Close()
}
