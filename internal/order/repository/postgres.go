package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"order-billing/internal/order/model"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

// три исхода:
// 1.Новый заказ: создаст и вернет order
// 2.Order уже есть: вернет пустой order{} и ErrAlreadyExist
// 3. системная ошибка
func (r *PostgresRepo) CreateOrder(ctx context.Context, orderID string, userID string, amount int64) (model.Order, bool, error) {
	var order model.Order

	const query = `
	INSERT INTO orders(id, user_id, amount)
	VALUES($1, $2, $3)
	ON CONFLICT (id) DO NOTHING
	RETURNING id, user_id, amount, status, fail_reason, created_at;
	`
	err := r.db.QueryRowContext(ctx, query, orderID, userID, amount).Scan(
		&order.ID,
		&order.UserID,
		&order.Amount,
		&order.Status,
		&order.FailReason,
		&order.CreatedAt,
	)
	//order создался
	if err == nil {
		return order, true, nil
	}

	//сис. ошибка
	if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("create order", "err", err.Error())
		return order, false, err
	}

	const selectQuery = `
	SELECT
	id, user_id, amount, status, fail_reason, created_at
	FROM orders
	WHERE id = $1;`

	if err := r.db.QueryRowContext(ctx, selectQuery, orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.Amount,
		&order.Status,
		&order.FailReason,
		&order.CreatedAt,
	); err != nil {
		slog.Error("select exist order", "err", err.Error())
		return model.Order{}, false, err
	}
	return order, false, nil
}

func (r *PostgresRepo) UpdateStatus(ctx context.Context, orderID string, status model.OrderStatus, failReason *string) (model.Order, error) {
	var order model.Order
	const query = `
	UPDATE orders
	SET status = $2, fail_reason = $3
	WHERE id = $1
	RETURNING id, user_id, amount, status, fail_reason, created_at;`

	if err := r.db.QueryRowContext(ctx, query, orderID, status, failReason).Scan(
		&order.ID,
		&order.UserID,
		&order.Amount,
		&order.Status,
		&order.FailReason,
		&order.CreatedAt,
	); err != nil {
		slog.Error(err.Error())
		return model.Order{}, err
	}
	return order, nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id string) (model.Order, error) {
	var order model.Order
	const query = `
	SELECT
	id, user_id, amount, status, fail_reason, created_at
	FROM orders
	WHERE id = $1;
	`

	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Amount,
		&order.Status,
		&order.FailReason,
		&order.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Order{}, model.ErrOrderNotFound
		}
		return model.Order{}, err
	}
	return order, nil
}
