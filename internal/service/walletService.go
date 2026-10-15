package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Homeveloper/PORestTZ/internal/domain"
)

type WalletRepository interface {
	ApplyOperation(ctx context.Context, walletID uuid.UUID, op domain.OperationType, amount decimal.Decimal) (decimal.Decimal, error)
	GetBalance(ctx context.Context, walletID uuid.UUID) (decimal.Decimal, error)
}

type WalletService struct {
	repo WalletRepository
}

func NewWalletService(repo WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) ProcessOperation(ctx context.Context, walletID uuid.UUID, op domain.OperationType, amount decimal.Decimal) (decimal.Decimal, error) {
	if !op.Valid() {
		return decimal.Zero, domain.ErrInvalidOperation
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, domain.ErrInvalidAmount
	}
	return s.repo.ApplyOperation(ctx, walletID, op, amount)
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (decimal.Decimal, error) {
	return s.repo.GetBalance(ctx, walletID)
}