package gapi

import (
	"context"
	"fmt"
	"github.com/fayca121/simplebank/token"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	//1- get access token from header
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		value := md.Get(authorizationHeaderKey)
		if len(value) == 0 {
			return nil, fmt.Errorf("authorization header is not provided")
		}
		authorizationHeaderValue := value[0]
		fields := strings.Fields(authorizationHeaderValue)
		//2- check if token is valid
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid authorization header format")
		}
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			return nil, fmt.Errorf("unsupported authorization type %s", authorizationType)
		}
		accessToken := fields[1]
		payload, err := server.tokenMaker.VerifyToken(accessToken)
		if err != nil {
			return nil, fmt.Errorf("invalid token: %s", err)
		}
		return payload, nil

	}
	return nil, fmt.Errorf("missing metada")
}
