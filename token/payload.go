package token

import (
	"github.com/google/uuid"
	"time"
)

type Payload struct {
	ID        string    `json:"id"`
	Issuer    string    `json:"issuer"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayLoad(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		Issuer:    issuer,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
		ID:        tokenID.String(),
	}

	return payload, nil
}
