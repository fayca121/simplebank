package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(queries *Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v", rbErr)
		}
	}

	return tx.Commit()
}

type TransferTxParam struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"fromAccount"`
	ToAccount   Account  `json:"toAccount"`
	FromEntry   Entry    `json:"fromEntry"`
	ToEntry     Entry    `json:"toEntry"`
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParam) (TransferTxResult, error) {

	var result TransferTxResult

	// function to be execute inside transaction
	err := store.execTx(ctx, func(q *Queries) error {
		//All operation for transfer amount from account1 to account2 (6 ops)
		var err error
		// 1- create transfer record
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})

		if err != nil {
			return err
		}

		// 2- create two entries records (from and to)
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})

		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID { //to avoid deadlock update should have the same direction 1 -> 2
			// Update source account's balance
			result.FromAccount, result.ToAccount, err = addMoney(ctx,
				q,
				arg.FromAccountID,
				-arg.Amount,
				arg.ToAccountID,
				arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx,
				q,
				arg.ToAccountID,
				arg.Amount,
				arg.FromAccountID,
				-arg.Amount)
		}

		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func addMoney(ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		Amount: amount1,
		ID:     accountID1,
	})
	if err != nil {
		return
	}
	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		Amount: amount2,
		ID:     accountID2,
	})
	if err != nil {
		return
	}
	return
}
