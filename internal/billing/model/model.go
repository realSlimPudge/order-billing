package model

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidID         = errors.New("invalid id")
)

type ChargeResult struct {
	NewBalance int64
}

type DepositResult struct {
	NewBalance int64
}
