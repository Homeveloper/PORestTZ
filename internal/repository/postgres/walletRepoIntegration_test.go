//go:build integration

package postgres

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/Homeveloper/PORestTZ/internal/domain"
)

func setupRepo(t *testing.T) *WalletRepo {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dsn := "postgres://postgres:postgres@localhost:5432/wallet?sslmode=disable"
	pool, err := NewPool(ctx, dsn)
	require.NoError(t, err, "postgres must be reachable (docker compose up -d postgres)")
	t.Cleanup(pool.Close)

	return NewWalletRepo(pool)
}


func TestConcurrentOperations(t *testing.T) {
	repo := setupRepo(t)
	ctx := context.Background()

	walletID := uuid.New()
	const (
		workers      = 50
		opsPerWorker = 20
		amount       = 100
	)


	_, err := repo.ApplyOperation(ctx, walletID, domain.OperationDeposit, decimal.NewFromInt(workers*opsPerWorker*amount))
	require.NoError(t, err)

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		netFlow     = decimal.Zero
		infraErrors int
	)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				op := domain.OperationDeposit
				if (worker+j)%2 == 0 {
					op = domain.OperationWithdraw
				}

				_, err := repo.ApplyOperation(ctx, walletID, op, decimal.NewFromInt(amount))

				mu.Lock()
				switch {
				case err == nil && op == domain.OperationDeposit:
					netFlow = netFlow.Add(decimal.NewFromInt(amount))
				case err == nil && op == domain.OperationWithdraw:
					netFlow = netFlow.Sub(decimal.NewFromInt(amount))
				case err == domain.ErrInsufficientFunds:
				default:
					infraErrors++
				}
				mu.Unlock()
			}
		}(w)
	}

	wg.Wait()

	require.Equal(t, 0, infraErrors, "ни одна операция не должна падать с инфраструктурной ошибкой")

	finalBalance, err := repo.GetBalance(ctx, walletID)
	require.NoError(t, err)

	expected := decimal.NewFromInt(workers * opsPerWorker * amount).Add(netFlow)
	require.True(t, finalBalance.Equal(expected), "потерянное обновление: ожидали %s, получили %s", expected, finalBalance)
}