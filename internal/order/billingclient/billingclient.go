package billingclient

import (
	"context"
	"order-billing/internal/order/model"
	"order-billing/proto/billing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	grpc billing.BillingServiceClient
}

func New(grpc billing.BillingServiceClient) *Client {
	return &Client{grpc: grpc}
}

func grpcToDomainErr(code codes.Code) error {
	switch code {
	case codes.InvalidArgument:
		return model.ErrValidationFailed
	case codes.NotFound:
		return model.ErrUserNotFound
	case codes.FailedPrecondition:
		return model.ErrDeclined
	default:
		return model.ErrBillingUnavailable
	}
}

func (c *Client) Charge(ctx context.Context, userID string, orderID string, amount int64) (int64, error) {
	req := &billing.ChargeRequest{UserId: userID, OrderId: orderID, Amount: amount}
	res, err := c.grpc.Charge(ctx, req)
	if err != nil {
		code := status.Code(err)
		domainErr := grpcToDomainErr(code)
		return 0, domainErr
	}
	balance := res.GetNewBalance()
	return balance, nil
}
