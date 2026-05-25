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
INSERT INTO transactions_ledger (idempotency_key, is_compensating_txn, transaction_id, account_id, other_party_account_id, amount_in_pennies)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: GetTransactionLedgerEntries :many
SELECT ledger.*, 
  other_party.account_holder_name AS other_party_name,
  txn.status
FROM transactions_ledger ledger
INNER JOIN accounts other_party ON ledger.other_party_account_id = other_party.id
INNER JOIN transactions txn ON ledger.transaction_id = txn.id
WHERE ledger.account_id = $1;

-- name: DuplicatePaymentCheck :one
SELECT id FROM transactions_ledger
WHERE account_id = $1 AND idempotency_key = $2
FOR UPDATE;
