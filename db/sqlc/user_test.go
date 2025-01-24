package db

import (
	"context"
	"database/sql"
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

func TestUpdateUserOnlyFullName(t *testing.T) {
	user := createRandomUser(t)

	newFullName := utils.RandomOwner()
	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: user.Username,
		FullName: sql.NullString{
			String: newFullName,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, user.FullName, newUser.FullName)
	require.Equal(t, newFullName, newUser.FullName)
	require.Equal(t, user.Email, newUser.Email)
	require.Equal(t, user.HashedPassword, newUser.HashedPassword)
}
func TestUpdateUserOnlyEmail(t *testing.T) {
	user := createRandomUser(t)

	newEmail := utils.RandomEmail()
	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: user.Username,
		Email: sql.NullString{
			String: newEmail,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, user.Email, newUser.Email)
	require.Equal(t, newEmail, newUser.Email)
	require.Equal(t, user.FullName, newUser.FullName)
	require.Equal(t, user.HashedPassword, newUser.HashedPassword)
}
func TestUpdateUserOnlyPassword(t *testing.T) {
	user := createRandomUser(t)

	newPassword, err := utils.HashPassword(utils.RandomString(6))

	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: user.Username,
		HashedPassword: sql.NullString{
			String: newPassword,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, user.HashedPassword, newUser.HashedPassword)
	require.Equal(t, newPassword, newUser.HashedPassword)
	require.Equal(t, user.FullName, newUser.FullName)
	require.Equal(t, user.Email, newUser.Email)
}

func TestUpdateUserAll(t *testing.T) {
	user := createRandomUser(t)

	newPassword, err := utils.HashPassword(utils.RandomString(6))
	newEmail := utils.RandomEmail()
	newFullName := utils.RandomOwner()

	newUser, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		Username: user.Username,
		HashedPassword: sql.NullString{
			String: newPassword,
			Valid:  true,
		},
		Email: sql.NullString{
			String: newEmail,
			Valid:  true,
		},
		FullName: sql.NullString{
			String: newFullName,
			Valid:  true,
		},
	})

	require.NoError(t, err)
	require.NotEqual(t, user.HashedPassword, newUser.HashedPassword)
	require.NotEqual(t, user.Email, newUser.Email)
	require.NotEqual(t, user.FullName, newUser.FullName)
	require.Equal(t, newPassword, newUser.HashedPassword)
	require.Equal(t, newFullName, newUser.FullName)
	require.Equal(t, newEmail, newUser.Email)
}
