package token

import (
	"errors"
	"time"
)

var (
	ErrInvalidToken = errors.New("token is not valid")
	ErrExpiredToken = errors.New("token has expired")
)

const issuer = "SimpleBank"

type Maker interface {
	CreateToken(username string, duration time.Duration) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}
