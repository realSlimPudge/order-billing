package model

import (
	"errors"
	"time"
)

var (
	ErrDeclined           = errors.New("order declined")
	ErrBillingUnavailable = errors.New("billing internal error")
	ErrUserNotFound       = errors.New("user not found")
	ErrValidationFailed   = errors.New("invalid payload")

	ErrOrderAlreadyExist = errors.New("order already exist")
	ErrOrderNotFound     = errors.New("order not found")
)

type OrderStatus string

const (
	PendingStatus OrderStatus = "pending"
	FailedStatus  OrderStatus = "failed"
	PaidStatus    OrderStatus = "paid"
)

type Order struct {
	ID         string
	UserID     string
	Amount     int64
	Status     OrderStatus
	FailReason *string
	CreatedAt  time.Time
}
