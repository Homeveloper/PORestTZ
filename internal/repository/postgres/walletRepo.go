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
	switch op {
	case domain.OperationDeposit:
		return r.deposit(ctx, walletID, amount)
	case domain.OperationWithdraw:
		return r.withdraw(ctx, walletID, amount)
	default:
		return decimal.Zero, domain.ErrInvalidOperation
	}
}


func (r *WalletRepo) deposit(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal) (decimal.Decimal, error) {
	var balance decimal.Decimal
	err := r.pool.QueryRow(ctx, `
		INSERT INTO wallets (id, balance) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET balance = wallets.balance + $2, updated_at = now()
		RETURNING balance
	`, walletID, amount).Scan(&balance)
	if err != nil {
		return decimal.Zero, fmt.Errorf("deposit: %w", err)
	}
	return balance, nil
}

func (r *WalletRepo) withdraw(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal) (decimal.Decimal, error) {
	var balance decimal.Decimal
	err := r.pool.QueryRow(ctx, `
		UPDATE wallets
		SET balance = balance - $2, updated_at = now()
		WHERE id = $1 AND balance >= $2
		RETURNING balance
	`, walletID, amount).Scan(&balance)

	if errors.Is(err, pgx.ErrNoRows) {
		
		var exists bool
		checkErr := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM wallets WHERE id = $1)`, walletID).Scan(&exists)
		if checkErr != nil {
			return decimal.Zero, fmt.Errorf("check wallet existence: %w", checkErr)
		}
		if !exists {
			return decimal.Zero, domain.ErrWalletNotFound
		}
		return decimal.Zero, domain.ErrInsufficientFunds
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("withdraw: %w", err)
	}
	return balance, nil
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