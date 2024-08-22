package utils

import (
	"crypto/rand"

	"github.com/oklog/ulid/v2"
)

func NewULID() ulid.ULID {
	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Now(), entropy)
	return id
}
