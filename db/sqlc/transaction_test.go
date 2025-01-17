package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valkyraycho/bank/utils"
)

func createRandomTransaction(t *testing.T, account1, account2 Account) Transaction {
	args := CreateTransactionParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        utils.RandomMoney(),
	}

	transaction, err := testQueries.CreateTransaction(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, transaction)

	require.Equal(t, args.FromAccountID, transaction.FromAccountID)
	require.Equal(t, args.ToAccountID, transaction.ToAccountID)
	require.Equal(t, args.Amount, transaction.Amount)

	require.NotZero(t, transaction.ID)
	require.NotZero(t, transaction.CreatedAt)
	return transaction
}

func TestCreateTransaction(t *testing.T) {
	createRandomTransaction(t, createRandomAccount(t), createRandomAccount(t))
}

func TestGetTransaction(t *testing.T) {
	transaction := createRandomTransaction(t, createRandomAccount(t), createRandomAccount(t))
	resTransaction, err := testQueries.GetTransaction(context.Background(), transaction.ID)

	require.NoError(t, err)
	require.NotEmpty(t, resTransaction)

	require.Equal(t, transaction.ID, resTransaction.ID)
	require.Equal(t, transaction.FromAccountID, resTransaction.FromAccountID)
	require.Equal(t, transaction.ToAccountID, resTransaction.ToAccountID)
	require.Equal(t, transaction.Amount, resTransaction.Amount)
	require.Equal(t, transaction.CreatedAt, resTransaction.CreatedAt)
}

func TestListTransactions(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	for i := 0; i < 5; i++ {
		createRandomTransaction(t, account1, account2)
		createRandomTransaction(t, account2, account1)
	}
	args := ListTransactionsParams{
		FromAccountID: account1.ID,
		ToAccountID:   account1.ID,
		Limit:         5,
		Offset:        5,
	}

	transactions, err := testQueries.ListTransactions(context.Background(), args)
	require.NoError(t, err)
	require.Len(t, transactions, 5)

	for _, transaction := range transactions {
		require.NotEmpty(t, transaction)
		require.True(t, transaction.FromAccountID == account1.ID || transaction.ToAccountID == account1.ID)
	}

}
