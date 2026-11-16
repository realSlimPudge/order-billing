package service

import (
	"context"
	"order-billing/internal/order/model"
)

type BillingClient interface {
	Charge(ctx context.Context, userID string, orderID string, amount int64) (int64, error)
}
type Repository interface {
	CreateOrder(ctx context.Context, orderID string, userId string, amount int64) (model.Order, bool, error)
	UpdateStatus(ctx context.Context, orderID string, status model.OrderStatus, failReason *string) (model.Order, error)
	GetByID(ctx context.Context, id string) (model.Order, error)
}

type OrderService struct {
	client BillingClient
	repo   Repository
}

func NewOrderService(client BillingClient, repo Repository) *OrderService {
	return &OrderService{client: client, repo: repo}
}

func (s *OrderService) attemptCharge(ctx context.Context, orderID string, userID string, amount int64) (model.Order, error) {
	_, err := s.client.Charge(ctx, userID, orderID, amount)
	status, reason := errToStatusReason(err)
	order, err := s.repo.UpdateStatus(ctx, orderID, status, reason)
	if err != nil {
		return model.Order{}, err
	}
	return order, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, orderID string, userID string, amount int64) (model.Order, bool, error) {
	if orderID == "" || userID == "" || amount <= 0 {
		return model.Order{}, false, model.ErrValidationFailed
	}

	order, created, err := s.repo.CreateOrder(ctx, orderID, userID, amount)
	//Сис. ошибка
	if err != nil {
		return model.Order{}, created, err
	}

	//Если существующий ордер со статусом paid/failed, то возращаем его
	if !created && order.Status != model.PendingStatus {
		return order, created, nil
	}

	//Если это новый ордер или существующий ордер со статусом pending, то оплата и смена статуса
	updatedOrder, err := s.attemptCharge(ctx, order.ID, order.UserID, order.Amount)
	if err != nil {
		return model.Order{}, created, err
	}
	return updatedOrder, created, nil
}

func (s *OrderService) GetByID(ctx context.Context, id string) (model.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.Order{}, err
	}
	return order, nil
}
