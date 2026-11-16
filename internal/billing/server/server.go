package server

import (
	"context"
	"errors"
	"fmt"
	"order-billing/internal/billing/model"
	"order-billing/internal/billing/service"
	"order-billing/proto/billing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	billing.UnimplementedBillingServiceServer
	svc *service.BillingService
}

func NewServer(svc *service.BillingService) *Server {
	return &Server{
		svc: svc,
	}
}

func (s *Server) Ping(ctx context.Context, req *billing.PingRequest) (*billing.PingResponse, error) {
	msg := req.GetMessage()

	newMsg := fmt.Sprintf("pong, msg: %s", msg)

	return &billing.PingResponse{Reply: newMsg}, nil
}

func (s *Server) Charge(ctx context.Context, req *billing.ChargeRequest) (*billing.ChargeResponse, error) {
	userID := req.GetUserId()
	orderID := req.GetOrderId()
	amount := req.GetAmount()

	result, err := s.svc.Charge(ctx, userID, orderID, amount)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidAmount),
			errors.Is(err, model.ErrInvalidID):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, model.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, model.ErrInsufficientFunds):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	return &billing.ChargeResponse{NewBalance: result.NewBalance}, nil
}
