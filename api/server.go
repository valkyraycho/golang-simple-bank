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

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", isValidCurrency)
	}

	server.registerRouter()
	return server, nil
}

func (s *Server) registerRouter() {
	router := gin.Default()

	router.POST("/users", s.createUser)
	router.POST("/users/login", s.loginUser)
	router.POST("/tokens/refresh", s.refreshAccessToken)

	authRouter := router.Group("/").Use(authMiddleware(s.tokenMaker))

	authRouter.GET("/accounts/:id", s.getAccount)
	authRouter.GET("/accounts", s.getAccounts)
	authRouter.POST("/accounts", s.createAccount)
	authRouter.POST("/transactions", s.createTransaction)
	s.router = router
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
