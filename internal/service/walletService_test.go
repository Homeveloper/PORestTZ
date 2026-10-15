package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/Homeveloper/PORestTZ/internal/domain"
)


type fakeRepo struct {
	applyFn   func(ctx context.Context, walletID uuid.UUID, op domain.OperationType, amount decimal.Decimal) (decimal.Decimal, error)
	balanceFn func(ctx context.Context, walletID uuid.UUID) (decimal.Decimal, error)
}

func (f *fakeRepo) ApplyOperation(ctx context.Context, walletID uuid.UUID, op domain.OperationType, amount decimal.Decimal) (decimal.Decimal, error) {
	return f.applyFn(ctx, walletID, op, amount)
}

func (f *fakeRepo) GetBalance(ctx context.Context, walletID uuid.UUID) (decimal.Decimal, error) {
	return f.balanceFn(ctx, walletID)
}

func TestProcessOperation_InvalidOperationType(t *testing.T) {
	svc := NewWalletService(&fakeRepo{})
	_, err := svc.ProcessOperation(context.Background(), uuid.New(), "UNKNOWN", decimal.NewFromInt(100))
	require.ErrorIs(t, err, domain.ErrInvalidOperation)
}

func TestProcessOperation_NonPositiveAmount(t *testing.T) {
	svc := NewWalletService(&fakeRepo{})

	_, err := svc.ProcessOperation(context.Background(), uuid.New(), domain.OperationDeposit, decimal.Zero)
	require.ErrorIs(t, err, domain.ErrInvalidAmount)

	_, err = svc.ProcessOperation(context.Background(), uuid.New(), domain.OperationDeposit, decimal.NewFromInt(-5))
	require.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestProcessOperation_InsufficientFunds(t *testing.T) {
	repo := &fakeRepo{
		applyFn: func(ctx context.Context, id uuid.UUID, op domain.OperationType, amount decimal.Decimal) (decimal.Decimal, error) {
			return decimal.Zero, domain.ErrInsufficientFunds
		},
	}

	svc := NewWalletService(repo)
	_, err := svc.ProcessOperation(context.Background(), uuid.New(), domain.OperationWithdraw, decimal.NewFromInt(100))
	require.ErrorIs(t, err, domain.ErrInsufficientFunds)
}

func TestGetBalance_NotFound(t *testing.T) {
	repo := &fakeRepo{
		balanceFn: func(ctx context.Context, id uuid.UUID) (decimal.Decimal, error) {
			return decimal.Zero, domain.ErrWalletNotFound
		},
	}

	svc := NewWalletService(repo)
	_, err := svc.GetBalance(context.Background(), uuid.New())
	require.ErrorIs(t, err, domain.ErrWalletNotFound)
}