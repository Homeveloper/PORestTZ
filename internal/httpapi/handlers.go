package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Homeveloper/PORestTZ/internal/domain"
)

type WalletService interface {
	ProcessOperation(ctx context.Context, walletID uuid.UUID, op domain.OperationType, amount decimal.Decimal) (decimal.Decimal, error)
	GetBalance(ctx context.Context, walletID uuid.UUID) (decimal.Decimal, error)
}

type Handler struct {
	svc WalletService
}

func NewHandler(svc WalletService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) PostWallet(w http.ResponseWriter, r *http.Request) {
	var req walletOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	walletID, err := uuid.Parse(req.WalletID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "valletId must be a valid UUID")
		return
	}

	newBalance, err := h.svc.ProcessOperation(r.Context(), walletID, req.OperationType, req.Amount)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, balanceResponse{
		WalletID: walletID.String(),
		Balance:  newBalance,
	})
}

func (h *Handler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "walletId")

	walletID, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "wallet id must be a valid UUID")
		return
	}

	balance, err := h.svc.GetBalance(r.Context(), walletID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, balanceResponse{
		WalletID: walletID.String(),
		Balance:  balance,
	})
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrWalletNotFound):
		writeError(w, http.StatusNotFound, "wallet not found")
	case errors.Is(err, domain.ErrInsufficientFunds):
		writeError(w, http.StatusBadRequest, "insufficient funds")
	case errors.Is(err, domain.ErrInvalidOperation):
		writeError(w, http.StatusBadRequest, "operationType must be DEPOSIT or WITHDRAW")
	case errors.Is(err, domain.ErrInvalidAmount):
		writeError(w, http.StatusBadRequest, "amount must be positive")
	default:
		log.Printf("internal error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}