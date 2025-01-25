package gapi

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/valkyraycho/bank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

func (s *Server) authorizeUser(ctx context.Context, accessibleRoles []string) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	auths := md.Get(authorizationHeader)
	if len(auths) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	fields := strings.Fields(auths[0])
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header")
	}

	authType := strings.ToLower(fields[0])
	if authType != authorizationBearer {
		return nil, fmt.Errorf("unsupported authorization type: %s", authType)
	}

	payload, err := s.tokenMaker.VerifyToken(fields[1])
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %s", err)
	}

	if !slices.Contains(accessibleRoles, payload.Role) {
		return nil, fmt.Errorf("permission denied")
	}
	return payload, nil
}
