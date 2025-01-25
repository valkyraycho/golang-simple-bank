package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/valkyraycho/bank/db/sqlc"
	"github.com/valkyraycho/bank/token"
)

type createTransactionRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (s *Server) createTransaction(ctx *gin.Context) {
	req := createTransactionRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := s.isValidAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != fromAccount.Owner {
		ctx.JSON(http.StatusUnauthorized, errors.New("from account doesn't belong to the authenticated user"))
		return
	}

	_, valid = s.isValidAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	res, err := s.store.TransferTx(ctx, db.TransactionTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (s *Server) isValidAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := s.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}
	if account.Currency != currency {
		ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)))
		return account, false
	}
	return account, true
}
