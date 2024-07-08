// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.21.0

package db

import (
	"database/sql"
	"time"
)

type User struct {
	ID            string
	Email         string
	FirstName     sql.NullString
	LastName      sql.NullString
	Password      string
	EmailVerified bool
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}