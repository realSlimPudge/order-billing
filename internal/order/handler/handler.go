package handler

import (
	"encoding/json"
	"net/http"
	"order-billing/internal/order/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var body CreateOrderReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	order, created, err := h.svc.CreateOrder(r.Context(), body.ID, body.UserID, body.Amount)
	if err != nil {
		respondError(w, httpStatusOf(err), err.Error())
		return
	}

	res := CreateOrderRes{
		ID:         order.ID,
		UserID:     order.UserID,
		Amount:     order.Amount,
		Status:     string(order.Status),
		FailReason: order.FailReason,
		CreatedAt:  order.CreatedAt,
	}

	if created {
		respondJSON(w, http.StatusCreated, res)
		return
	}
	respondJSON(w, http.StatusOK, res)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	order, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, httpStatusOf(err), err.Error())
		return
	}

	res := CreateOrderRes{
		ID:         order.ID,
		UserID:     order.UserID,
		Amount:     order.Amount,
		Status:     string(order.Status),
		FailReason: order.FailReason,
		CreatedAt:  order.CreatedAt,
	}

	respondJSON(w, http.StatusOK, res)
}
