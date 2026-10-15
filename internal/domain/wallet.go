package domain

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OperationType string

const (
	OperationDeposit  OperationType = "DEPOSIT"
	OperationWithdraw OperationType = "WITHDRAW"
)

func (o OperationType) Valid() bool {
	return o == OperationDeposit || o == OperationWithdraw
}

type Wallet struct {
	ID      uuid.UUID
	Balance decimal.Decimal
}

var (
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidOperation  = errors.New("invalid operation type")
	ErrInvalidAmount     = errors.New("amount must be positive")
)