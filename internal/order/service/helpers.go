package service

import (
	"errors"
	"order-billing/internal/order/model"
)

func errToStatusReason(err error) (model.OrderStatus, *string) {
	switch {
	case err == nil:
		return model.PaidStatus, nil
	case errors.Is(err, model.ErrBillingUnavailable):
		return model.PendingStatus, nil
	case errors.Is(err, model.ErrValidationFailed),
		errors.Is(err, model.ErrUserNotFound),
		errors.Is(err, model.ErrDeclined):
		reason := err.Error()
		return model.FailedStatus, &reason
	default:
		return model.PendingStatus, nil
	}
}
