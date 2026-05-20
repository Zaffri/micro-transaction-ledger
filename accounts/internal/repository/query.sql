-- name: CreateAccount :one
INSERT INTO accounts (balance_in_pennies, account_holder_name)
VALUES ($1, $2)
RETURNING id, balance_in_pennies;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1;

-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1
FOR UPDATE;

-- name: UpdateBalance :exec
UPDATE accounts
SET balance_in_pennies = $2,
    updated_at = $3
WHERE id = $1;
