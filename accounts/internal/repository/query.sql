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
SET balance_in_pennies = $2
WHERE id = $1;

-- name: CreateTransaction :one
INSERT INTO transactions (description, account_id)
VALUES ($1, $2)
RETURNING id, status;

-- name: UpdateTransaction :exec
UPDATE transactions
SET status = $2
WHERE id = $1;

-- name: CreateTransactionLedgerEntry :one
INSERT INTO transactions_ledger (transaction_id, account_id, other_party_account_id, amount_in_pennies)
VALUES ($1, $2, $3, $4)
RETURNING id;

