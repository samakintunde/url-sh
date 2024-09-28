package utils

import (
	"crypto/rand"
	"errors"
	"strings"
)

var (
	ErrGeneratingToken = errors.New("couldn't generate token")
)

const tokenChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed 0, O, 1, I

func GenerateAlphanum(length int) (string, error) {
	tokenLength := 8
	if length != 0 {
		tokenLength = length
	}
	bytes := make([]byte, tokenLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	for _, b := range bytes {
		result.WriteByte(tokenChars[int(b)%len(tokenChars)])
	}

	if result.Len() != tokenLength {
		return "", errors.New("generated token has unexpected length")
	}

	return result.String(), nil
}
