package token

import (
	"github.com/fayca121/simplebank/util"
	"github.com/google/uuid"
	"time"
)

type Payload struct {
	ID        string    `json:"id"`
	Issuer    string    `json:"issuer"`
	Role      string    `json:"role"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewPayLoad(username string, role util.Role, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		Issuer:    issuer,
		Username:  username,
		Role:      role.String(),
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
		ID:        tokenID.String(),
	}

	return payload, nil
}
