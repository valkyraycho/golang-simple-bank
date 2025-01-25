package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/valkyraycho/bank/utils"
)

func createRandomAccount(t *testing.T) Account {
	user := createRandomUser(t)
	args := CreateAccountParams{
		Owner:    user.Username,
		Balance:  utils.RandomMoney(),
		Currency: utils.RandomCurrency(),
	}

	account, err := testStore.CreateAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, args.Owner, account.Owner)
	require.Equal(t, args.Balance, account.Balance)
	require.Equal(t, args.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	account := createRandomAccount(t)
	resAccount, err := testStore.GetAccount(context.Background(), account.ID)

	require.NoError(t, err)
	require.NotEmpty(t, resAccount)

	require.Equal(t, account.ID, resAccount.ID)
	require.Equal(t, account.Owner, resAccount.Owner)
	require.Equal(t, account.Balance, resAccount.Balance)
	require.Equal(t, account.Currency, resAccount.Currency)
	require.Equal(t, account.CreatedAt, resAccount.CreatedAt)
}

func TestUpdateAccount(t *testing.T) {
	account := createRandomAccount(t)

	args := UpdateAccountParams{
		ID:      account.ID,
		Balance: utils.RandomMoney(),
	}

	resAccount, err := testStore.UpdateAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, resAccount)

	require.Equal(t, account.ID, resAccount.ID)
	require.Equal(t, account.Owner, resAccount.Owner)
	require.Equal(t, args.Balance, resAccount.Balance)
	require.Equal(t, account.Currency, resAccount.Currency)
	require.Equal(t, account.CreatedAt, resAccount.CreatedAt)
}

func TestDeleteAccount(t *testing.T) {
	account := createRandomAccount(t)

	err := testStore.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)
	deletedAccount, err := testStore.GetAccount(context.Background(), account.ID)
	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, deletedAccount)
}

func TestListAccounts(t *testing.T) {
	lastAccount := Account{}

	for i := 0; i < 10; i++ {
		lastAccount = createRandomAccount(t)
	}

	args := ListAccountsParams{
		Owner:  lastAccount.Owner,
		Limit:  5,
		Offset: 0,
	}

	accounts, err := testStore.ListAccounts(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.Equal(t, lastAccount.Owner, account.Owner)
	}

}
