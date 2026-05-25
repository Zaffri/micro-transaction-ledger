-- name: CreateTransactionLedgerEntry :one
INSERT INTO transactions_ledger (transaction_id, account_id, other_party_account_id, idempotency_key, is_compensating_txn, amount_in_pennies)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;
