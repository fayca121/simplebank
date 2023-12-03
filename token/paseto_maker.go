package token

import (
	"aidanwoods.dev/go-paseto"
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

func (p *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayLoad(username, duration)
	if err != nil {
		return "", err
	}

	// create Token
	token := paseto.NewToken()
	token.SetSubject(payload.Username)
	token.SetIssuer(payload.Issuer)
	token.SetIssuedAt(payload.IssuedAt)
	token.SetExpiration(payload.ExpiredAt)
	err = token.Set("ID", payload.ID)

	if err != nil {
		return "", err
	}
	encrypted := token.V4Encrypt(p.symmetricKey, nil)
	return encrypted, nil
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
	var id string
	err = verifiedToken.Get("ID", &id)

	if err != nil {
		return nil, err
	}

	return &Payload{
		Issuer:    issuer,
		Username:  subject,
		ExpiredAt: expiredAt,
		IssuedAt:  issuedAt,
		ID:        id,
	}, nil
}
