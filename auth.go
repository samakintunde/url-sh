package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	hibp "github.com/mattevans/pwned-passwords"
	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/argon2"
)

type ArgonParams struct {
	saltLength  uint32
	memory      uint32
	keyLength   uint32
	iterations  uint32
	parallelism uint8
}

var argonParams = &ArgonParams{
	saltLength:  16,
	memory:      20 * 1024,
	keyLength:   32,
	iterations:  2,
	parallelism: 1,
}

func generateSalt(n int32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func HashPassword(password string) (string, error) {
	salt, err := generateSalt(int32(argonParams.saltLength))
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonParams.iterations, argonParams.memory, argonParams.parallelism, argonParams.keyLength)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonParams.memory, argonParams.iterations, argonParams.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))

	return encodedHash, nil
}

func checkPasswordStrength(password string) bool {
	strength := zxcvbn.PasswordStrength(password, nil)
	return strength.Score >= 3
}

func checkPasswordPwned(password string) (bool, error) {
	client := hibp.NewClient()
	compromised, err := client.Compromised(password)
	if err != nil {
		return false, err
	}
	return compromised, nil
}
