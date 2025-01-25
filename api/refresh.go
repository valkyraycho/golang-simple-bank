package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type refreshAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type refreshAccessTokenResponse struct {
	AccessToken          string `json:"access_token"`
	AccessTokenExpiresAt string `json:"access_token_expires_at"`
}

func (s *Server) refreshAccessToken(ctx *gin.Context) {
	req := refreshAccessTokenRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	refreshPayload, err := s.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	session, err := s.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	if session.IsBlocked {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("blocked session")))
		return
	}

	if session.Username != refreshPayload.Username {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("incorrect session user")))
		return
	}

	if session.RefreshToken != req.RefreshToken {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("mismatch session token")))
		return
	}

	if time.Now().After(session.ExpiresAt) {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("expired session")))
		return
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(refreshPayload.Username, refreshPayload.Role, s.cfg.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, refreshAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt.String(),
	})
	return
}
