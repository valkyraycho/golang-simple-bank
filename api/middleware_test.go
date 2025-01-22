package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/valkyraycho/bank/token"
)

func addAuthorizationToRequest(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	accessToken, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	request.Header.Set(authorizationHeaderKey, fmt.Sprintf("%s %s", authorizationType, accessToken))
}
func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationToRequest(
					t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					"user",
					time.Minute,
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "NoAuth",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnsupportedAuth",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationToRequest(
					t,
					request,
					tokenMaker,
					"unsupported",
					"user",
					time.Minute,
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationToRequest(
					t,
					request,
					tokenMaker,
					"",
					"user",
					time.Minute,
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorizationToRequest(
					t,
					request,
					tokenMaker,
					authorizationTypeBearer,
					"user",
					-time.Minute,
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, testCase := range testCases {
		server := NewTestServer(t, nil)

		authPath := "/auth"
		server.router.GET(
			authPath,
			authMiddleware(server.tokenMaker),
			func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{})
			},
		)

		recorder := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, authPath, nil)
		require.NoError(t, err)

		testCase.setupAuth(t, req, server.tokenMaker)
		server.router.ServeHTTP(recorder, req)
		testCase.checkResponse(t, recorder)
	}
}
