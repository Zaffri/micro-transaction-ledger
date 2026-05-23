-- name: CreateTransactionLedgerEntry :one
INSERT INTO transactions_ledger (transaction_id, account_id, other_party_account_id, amount_in_pennies)
VALUES ($1, $2, $3, $4)
RETURNING id;
