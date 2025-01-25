package gapi

import (
	"fmt"

	db "github.com/valkyraycho/bank/db/sqlc"
	"github.com/valkyraycho/bank/pb"
	"github.com/valkyraycho/bank/token"
	"github.com/valkyraycho/bank/utils"
	"github.com/valkyraycho/bank/worker"
)

type Server struct {
	pb.UnimplementedBankServiceServer
	cfg             utils.Config
	store           db.Store
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

func NewServer(cfg utils.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(cfg.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{cfg: cfg, store: store, tokenMaker: tokenMaker, taskDistributor: taskDistributor}

	return server, nil
}
