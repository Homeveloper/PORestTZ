package httpapi

import (
	"github.com/shopspring/decimal"

	"github.com/Homeveloper/PORestTZ/internal/domain"
)

type walletOperationRequest struct {
	WalletID      string               `json:"valletId"`
	OperationType domain.OperationType `json:"operationType"`
	Amount        decimal.Decimal      `json:"amount"`
}

type balanceResponse struct {
	WalletID string          `json:"walletId"`
	Balance  decimal.Decimal `json:"balance"`
}

type errorResponse struct {
	Error string `json:"error"`
}