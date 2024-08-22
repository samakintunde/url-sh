package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

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

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, fmt.Errorf("incompatible version")
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &argonParams.memory, &argonParams.iterations, &argonParams.parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	argonParams.keyLength = uint32(len(decodedHash))

	comparisonHash := argon2.IDKey([]byte(password), salt, argonParams.iterations, argonParams.memory, argonParams.parallelism, argonParams.keyLength)

	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}

func CheckPasswordStrength(password string) bool {
	strength := zxcvbn.PasswordStrength(password, nil)
	return strength.Score >= 3
}

func CheckPasswordPwned(password string) (bool, error) {
	client := hibp.NewClient()
	compromised, err := client.Compromised(password)
	if err != nil {
		return false, err
	}
	return compromised, nil
}
