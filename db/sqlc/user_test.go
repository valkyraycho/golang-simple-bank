package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valkyraycho/bank/utils"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := utils.HashPassword(utils.RandomString(6))
	require.NoError(t, err)

	args := CreateUserParams{
		Username:       utils.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       utils.RandomOwner(),
		Email:          utils.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, args.Username, user.Username)
	require.Equal(t, args.FullName, user.FullName)
	require.Equal(t, args.HashedPassword, user.HashedPassword)
	require.Equal(t, args.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)
	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user := createRandomUser(t)
	resUser, err := testQueries.GetUser(context.Background(), user.Username)

	require.NoError(t, err)
	require.NotEmpty(t, resUser)

	require.Equal(t, user.Username, resUser.Username)
	require.Equal(t, user.HashedPassword, resUser.HashedPassword)
	require.Equal(t, user.FullName, resUser.FullName)
	require.Equal(t, user.Email, resUser.Email)
	require.Equal(t, user.CreatedAt, resUser.CreatedAt)
	require.Equal(t, user.PasswordChangedAt, resUser.PasswordChangedAt)
}
