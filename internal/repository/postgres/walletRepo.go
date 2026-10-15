package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/Homeveloper/PORestTZ/internal/domain"
)

type WalletRepo struct {
	pool *pgxpool.Pool
}

func NewWalletRepo(pool *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{pool: pool}
}


func (r *WalletRepo) ApplyOperation(ctx context.Context, walletID uuid.UUID, op domain.OperationType, amount decimal.Decimal) (decimal.Decimal, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback(ctx)

	if op == domain.OperationDeposit {
		_, err = tx.Exec(ctx, `
			INSERT INTO wallets (id, balance) VALUES ($1, 0)
			ON CONFLICT (id) DO NOTHING
		`, walletID)
		if err != nil {
			return decimal.Zero, fmt.Errorf("ensure wallet: %w", err)
		}
	}

	var balance decimal.Decimal
	err = tx.QueryRow(ctx, `SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`, walletID).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return decimal.Zero, domain.ErrWalletNotFound
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("lock wallet: %w", err)
	}

	var newBalance decimal.Decimal
	switch op {
	case domain.OperationDeposit:
		newBalance = balance.Add(amount)
	case domain.OperationWithdraw:
		if balance.LessThan(amount) {
			return decimal.Zero, domain.ErrInsufficientFunds
		}
		newBalance = balance.Sub(amount)
	default:
		return decimal.Zero, domain.ErrInvalidOperation
	}

	_, err = tx.Exec(ctx, `UPDATE wallets SET balance = $1, updated_at = now() WHERE id = $2`, newBalance, walletID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("update balance: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return decimal.Zero, fmt.Errorf("commit tx: %w", err)
	}

	return newBalance, nil
}

func (r *WalletRepo) GetBalance(ctx context.Context, walletID uuid.UUID) (decimal.Decimal, error) {
	var balance decimal.Decimal
	err := r.pool.QueryRow(ctx, `SELECT balance FROM wallets WHERE id = $1`, walletID).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return decimal.Zero, domain.ErrWalletNotFound
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("get balance: %w", err)
	}
	return balance, nil
}