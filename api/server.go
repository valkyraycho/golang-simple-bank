package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/valkyraycho/bank/db/sqlc"
	"github.com/valkyraycho/bank/token"
	"github.com/valkyraycho/bank/utils"
)

type Server struct {
	cfg        utils.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(cfg utils.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(cfg.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{cfg: cfg, store: store, tokenMaker: tokenMaker}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", isValidCurrency)
	}

	router.POST("/users", server.createUser)

	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.getAccounts)
	router.POST("/accounts", server.createAccount)
	router.POST("/transactions", server.createTransaction)

	server.router = router
	return server, nil
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
