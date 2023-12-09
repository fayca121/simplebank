package token

import (
	"fmt"
	"github.com/fayca121/simplebank/util"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

type jwtClaims struct {
	*jwt.RegisteredClaims
	Role string `json:"role,omitempty"`
}

func NewJWTMaker(secretKey string) (*JWTMaker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{
		secretKey: secretKey,
	}, nil
}

func (maker *JWTMaker) CreateToken(username string, role util.Role, duration time.Duration) (string, *Payload, error) {

	payload, err := NewPayLoad(username, role, duration)

	if err != nil {
		return "", nil, err
	}
	//create claims from payload
	claims := payLoadToJWTClaims(payload)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	if err != nil {
		return "", nil, err
	}
	return token, payload, nil

}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	t, err := jwt.ParseWithClaims(token, &jwtClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, ErrInvalidToken
			}
			return []byte(maker.secretKey), nil
		})

	if err != nil {
		return nil, err
	}

	if !t.Valid {
		return nil, ErrInvalidToken
	}

	payload, err := jwtClaimsToPayLoad(t.Claims)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

// mapping functions
func payLoadToJWTClaims(payload *Payload) jwt.Claims {
	return &jwtClaims{
		RegisteredClaims: &jwt.RegisteredClaims{
			Issuer:    payload.Issuer,
			Subject:   payload.Username,
			ExpiresAt: jwt.NewNumericDate(payload.ExpiredAt),
			IssuedAt:  jwt.NewNumericDate(payload.IssuedAt),
			ID:        payload.ID,
		},
		Role: payload.Role,
	}
}

func jwtClaimsToPayLoad(c jwt.Claims) (*Payload, error) {
	claims, ok := c.(*jwtClaims)
	if !ok {
		return nil, fmt.Errorf("cannot retreive claims data from token")
	}
	return &Payload{
		ID:        claims.ID,
		Issuer:    claims.Issuer,
		Role:      claims.Role,
		Username:  claims.Subject,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiredAt: claims.ExpiresAt.Time,
	}, nil
}
