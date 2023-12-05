package api

import (
	"fmt"
	db "github.com/fayca121/simplebank/db/sqlc"
	"github.com/fayca121/simplebank/token"
	"github.com/fayca121/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
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
	//add routes to router
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("currency", validCurrency)
	}
	accountGrp := router.Group("/accounts").Use(authMiddleware(tokenMaker))
	{
		accountGrp.POST("/", server.createAccount)
		accountGrp.GET("/:id", server.getAccount)
		accountGrp.GET("/", server.listAccount)
		accountGrp.PUT("/", server.updateAccount)
		accountGrp.DELETE("/:id", server.deleteAccount)
	}

	router.POST("/transfers", server.createTransfer)
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)
	server.router = router
	return server, nil
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}
