package handler

import "time"

type CreateOrderReq struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Amount int64  `json:"amount"`
}

type CreateOrderRes struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Amount     int64     `json:"amount"`
	Status     string    `json:"status"`
	FailReason *string   `json:"fail_reason,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
