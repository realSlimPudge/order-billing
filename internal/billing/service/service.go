package service

import (
	"context"
	"order-billing/internal/billing/model"
)

type Repository interface {
	Charge(ctx context.Context, userID string, orderID string, amount int64) (int64, error)
	Deposit(ctx context.Context, userID string, amount int64) (int64, error)
}

type BillingService struct {
	repo Repository
}

func NewBillingService(repo Repository) *BillingService {
	return &BillingService{
		repo: repo,
	}
}

func (s *BillingService) Charge(ctx context.Context, userID string, orderID string, amount int64) (model.ChargeResult, error) {
	if userID == "" || orderID == "" {
		return model.ChargeResult{}, model.ErrInvalidID
	}
	if amount <= 0 {
		return model.ChargeResult{}, model.ErrInvalidAmount
	}

	newBalance, err := s.repo.Charge(ctx, userID, orderID, amount)
	if err != nil {
		return model.ChargeResult{}, err
	}
	result := model.ChargeResult{NewBalance: newBalance}

	return result, nil
}

func (s *BillingService) Deposit(ctx context.Context, userID string, amount int64) (model.DepositResult, error) {
	if amount <= 0 {
		return model.DepositResult{}, model.ErrInvalidAmount
	}
	newBalance, err := s.repo.Deposit(ctx, userID, amount)
	if err != nil {
		return model.DepositResult{}, err
	}
	result := model.DepositResult{NewBalance: newBalance}

	return result, nil
}
