package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	db "github.com/valkyraycho/bank/db/sqlc"
	"github.com/valkyraycho/bank/token"
	"github.com/valkyraycho/bank/utils"
	"github.com/valkyraycho/bank/worker"
	"google.golang.org/grpc/metadata"
)

func NewTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	cfg := utils.Config{
		TokenSymmetricKey:   utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	server, err := NewServer(cfg, store, taskDistributor)
	require.NoError(t, err)
	return server
}

func newContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, role string, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)
	md := metadata.MD{
		authorizationHeader: []string{
			fmt.Sprintf("%s %s", authorizationBearer, accessToken),
		},
	}
	return metadata.NewIncomingContext(context.Background(), md)

}
