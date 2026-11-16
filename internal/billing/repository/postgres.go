package repository

import (
	"context"
	"database/sql"
	"errors"
	"order-billing/internal/billing/model"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) Charge(ctx context.Context, userID string, orderID string, amount int64) (int64, error) {
	//Открываем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	//Проверяем существует ли user, если да - получаем баланс, нет - ошибка ErrUserNotFound
	var userBalance int64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT balance FROM accounts WHERE user_id = $1 FOR UPDATE`,
		userID,
	).Scan(&userBalance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, model.ErrUserNotFound
		}
		return 0, err
	}

	//Проверяем идентпотентность
	var newBalance int64
	err = tx.QueryRowContext(
		ctx,
		`SELECT new_balance FROM charges WHERE order_id = $1`,
		orderID,
	).Scan(&newBalance)

	//Запись найдена - возвращаем из журнала(реализация идентпотентности)
	if err == nil {
		tx.Commit()
		return newBalance, nil
	}

	//Любая ошибка кроме не найдено, тоесть системная, возвращаем наружу
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	//Проверка на баланс
	if userBalance < amount {
		return 0, model.ErrInsufficientFunds
	}

	var balanceAfterChange int64
	//Ставим новый баланс для аккаунта (balance - amount)
	err = tx.QueryRowContext(
		ctx,
		`UPDATE accounts
		SET balance=balance-$1
	 	WHERE user_id=$2
		RETURNING balance`,
		amount, userID,
	).Scan(&balanceAfterChange)
	if err != nil {
		return 0, err
	}

	//Создаем новую запись в журнал
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO charges(order_id, user_id, amount, new_balance)
	 	VALUES ($1, $2, $3, $4)`,
		orderID, userID, amount, balanceAfterChange,
	)
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return balanceAfterChange, nil
}

func (r *PostgresRepo) Deposit(ctx context.Context, userID string, amount int64) (int64, error) {
	var newBalance int64
	if err := r.db.QueryRowContext(
		ctx,
		`
		INSERT INTO accounts (user_id, balance)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET
			balance = accounts.balance + EXCLUDED.balance
		RETURNING balance;
		`,
		userID, amount,
	).Scan(&newBalance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, model.ErrUserNotFound
		}
		return 0, err
	}
	return newBalance, nil
}
