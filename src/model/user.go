package model

import (
	"time"
)

// User of l10n-center
type User struct {
	Email      string     `db:"email"`
	Passhash   []byte     `db:"passhash"`
	Role       Role       `db:"role"`
	ResetToken []byte     `db:"reset_token"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}
