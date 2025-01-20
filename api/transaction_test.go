package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	mock_db "github.com/valkyraycho/bank/db/mock"
	db "github.com/valkyraycho/bank/db/sqlc"
	"go.uber.org/mock/gomock"
)

func TestCreateTransaction(t *testing.T) {
	amount := int64(10)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": 1,
				"to_account_id":   2,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(int64(1))).
					Times(1).
					Return(db.Account{
						ID:       1,
						Currency: "USD",
					}, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(int64(2))).
					Times(1).
					Return(db.Account{
						ID:       2,
						Currency: "USD",
					}, nil)
				arg := db.CreateTransactionParams{
					FromAccountID: 1,
					ToAccountID:   2,
					Amount:        amount,
				}
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"from_account_id": 1,
				"to_account_id":   2,
				"amount":          amount,
				"currency":        "XYZ",
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NegativeAmount",
			body: gin.H{
				"from_account_id": 1,
				"to_account_id":   2,
				"amount":          -amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": 1,
				"to_account_id":   2,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(int64(1))).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(int64(2))).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "CurrencyMismatch",
			body: gin.H{
				"from_account_id": 1,
				"to_account_id":   2,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(int64(1))).Times(1).Return(db.Account{
					ID:       1,
					Currency: "EUR",
				}, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(int64(2))).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_db.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/transactions"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
