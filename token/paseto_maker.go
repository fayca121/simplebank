package token

import (
	"aidanwoods.dev/go-paseto"
	"github.com/fayca121/simplebank/util"
	"time"
)

type PasetoMaker struct {
	symmetricKey paseto.V4SymmetricKey
}

func NewPasetoMaker() (*PasetoMaker, error) {
	return &PasetoMaker{
		symmetricKey: paseto.NewV4SymmetricKey(),
	}, nil
}

func (p *PasetoMaker) CreateToken(username string, role util.Role, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayLoad(username, role, duration)
	if err != nil {
		return "", nil, err
	}

	// create Token
	token := paseto.NewToken()
	token.SetSubject(payload.Username)
	token.SetIssuer(payload.Issuer)
	token.SetIssuedAt(payload.IssuedAt)
	token.SetExpiration(payload.ExpiredAt)
	token.SetString("ID", payload.ID)
	token.SetString("role", payload.Role)

	encrypted := token.V4Encrypt(p.symmetricKey, nil)
	return encrypted, payload, nil
}

func (p *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	verifiedToken, err := paseto.NewParser().ParseV4Local(p.symmetricKey, token, nil)
	if err != nil {
		return nil, err
	}
	issuer, err := verifiedToken.GetIssuer()
	if err != nil {
		return nil, err
	}
	subject, err := verifiedToken.GetSubject()
	if err != nil {
		return nil, err
	}

	issuedAt, err := verifiedToken.GetIssuedAt()
	if err != nil {
		return nil, err
	}
	expiredAt, err := verifiedToken.GetExpiration()
	if err != nil {
		return nil, err
	}
	id, err := verifiedToken.GetString("ID")

	if err != nil {
		return nil, err
	}
	role, err := verifiedToken.GetString("role")

	if err != nil {
		return nil, err
	}

	return &Payload{
		Issuer:    issuer,
		Role:      role,
		Username:  subject,
		ExpiredAt: expiredAt,
		IssuedAt:  issuedAt,
		ID:        id,
	}, nil
}
