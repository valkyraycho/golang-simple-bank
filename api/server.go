package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/valkyraycho/bank/db/sqlc"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.getAccounts)
	router.POST("/accounts", server.createAccount)

	server.router = router
	return server
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
