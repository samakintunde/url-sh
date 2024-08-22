package token

import (
	"encoding/json"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

type PasetoMaker struct {
	symmetricKey paseto.V4SymmetricKey
}

func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("SymmetricKey too short should be: %v", chacha20poly1305.KeySize)
	}

	v4SymmetricKey, err := paseto.V4SymmetricKeyFromBytes([]byte(symmetricKey))

	if err != nil {
		return nil, err
	}

	return &PasetoMaker{
		symmetricKey: v4SymmetricKey,
	}, nil
}

func (maker *PasetoMaker) CreateToken(userID string, duration time.Duration) (string, *Claims, error) {
	claims := NewClaims(userID, duration)

	token := paseto.NewToken()

	token.Set("user_id", claims.UserID)
	token.SetSubject(claims.UserID)
	token.SetExpiration(claims.ExpiresAt)
	token.SetIssuedAt(claims.IssuedAt)

	encryptedToken := token.V4Encrypt(maker.symmetricKey, nil)

	return encryptedToken, claims, nil
}

func (maker *PasetoMaker) VerifyToken(encryptedToken string) (*Claims, error) {
	claims := &Claims{}

	parser := paseto.NewParser()
	token, err := parser.ParseV4Local(maker.symmetricKey, encryptedToken, nil)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(token.ClaimsJSON(), claims)

	if err != nil {
		return nil, err
	}

	return claims, nil
}
