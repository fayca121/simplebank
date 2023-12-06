package gapi

import (
	"fmt"
	db "github.com/fayca121/simplebank/db/sqlc"
	"github.com/fayca121/simplebank/pb"
	"github.com/fayca121/simplebank/token"
	"github.com/fayca121/simplebank/util"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey) // or token.NewPasetaMaker()

	if err != nil {
		return nil, fmt.Errorf("cannot create token maker")
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
