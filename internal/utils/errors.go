package utils

import "github.com/mattn/go-sqlite3"

func IsConflictError(err error) bool {
	if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
		return true
	}
	return false
}
