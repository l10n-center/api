package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// User of l10n-center
//
// nolint: aligncheck
type User struct {
	ID         bson.ObjectId `bson:"_id"`
	Email      string        `bson:"email"`
	Passhash   []byte        `bson:"passhash"`
	ResetToken []byte        `bson:"resetToken"`
	IsAdmin    bool          `bson:"isAdmin"`
	Permission Permission    `bson:"permission"`
	CreatedAt  time.Time     `bson:"createdAt"`
	UpdatedAt  time.Time     `bson:"updatedAt"`
	DeletedAt  *time.Time    `bson:"deletedAt"`
}
