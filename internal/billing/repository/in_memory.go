package repository

import (
	"context"
	"order-billing/internal/billing/model"
	"sync"
)

type InMemoryRepo struct {
	balances  map[string]int64
	processed map[string]int64
	mu        sync.Mutex
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		balances:  make(map[string]int64),
		processed: make(map[string]int64),
	}
}

func (r *InMemoryRepo) Charge(ctx context.Context, userID string, orderID string, amount int64) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	newBalance, ok := r.processed[orderID]
	if ok {
		return newBalance, nil
	}

	userBalance, ok := r.balances[userID]
	if !ok {
		return 0, model.ErrUserNotFound
	}

	if userBalance < amount {
		return 0, model.ErrInsufficientFunds
	}

	newBalance = userBalance - amount
	r.balances[userID] = newBalance
	r.processed[orderID] = newBalance
	return newBalance, nil
}

func (r *InMemoryRepo) Deposit(ctx context.Context, userID string, amount int64) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.balances[userID] += amount

	return r.balances[userID], nil
}
