CREATE TABLE accounts(
    user_id TEXT PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE charges(
    order_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES accounts(user_id),
    amount BIGINT NOT NULL,
    new_balance BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
