-- name: CreateTransaction :one
INSERT INTO transactions (description, account_id, status)
VALUES ($1, $2, $3)
RETURNING id, status;
