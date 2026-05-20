-- name: CreateFraudCheck :one
INSERT INTO fraud_checks (transaction_id, risk_score)
VALUES ($1, $2)
RETURNING id;
