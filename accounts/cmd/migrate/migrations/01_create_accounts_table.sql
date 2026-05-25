CREATE EXTENSION IF NOT EXISTS moddatetime;

CREATE TABLE IF NOT EXISTS accounts (
  id BIGSERIAL PRIMARY KEY,
  balance_in_pennies BIGINT DEFAULT 0 NOT NULL,
  account_holder_name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trigger_update_accounts_timestamp ON accounts;

CREATE TRIGGER trigger_update_accounts_timestamp
  BEFORE UPDATE ON accounts
  FOR EACH ROW
  EXECUTE PROCEDURE moddatetime(updated_at);

-- DROP TYPE IF EXISTS transaction_status CASCADE; -- for dev only, can delete data

CREATE TYPE transaction_status AS ENUM ('pending', 'settled', 'rejected_fraud');

CREATE TABLE IF NOT EXISTS transactions (
  id BIGSERIAL PRIMARY KEY,
  status transaction_status NOT NULL DEFAULT 'pending',
  description TEXT NOT NULL,
  account_id BIGINT NOT NULL REFERENCES accounts(id) on DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trigger_update_transactions_timestamp ON transactions;

CREATE TRIGGER trigger_update_transactions_timestamp
  BEFORE UPDATE ON transactions
  FOR EACH ROW
  EXECUTE PROCEDURE moddatetime(updated_at);

CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);

CREATE TABLE IF NOT EXISTS transactions_ledger (
  id BIGSERIAL PRIMARY KEY,
  transaction_id BIGINT NOT NULL REFERENCES transactions(id) on DELETE RESTRICT,
  account_id BIGINT NOT NULL REFERENCES accounts(id) on DELETE RESTRICT,
  other_party_account_id BIGINT NOT NULL REFERENCES accounts(id) on DELETE RESTRICT, -- the other account involved
  idempotency_key UUID NOT NULL,
  is_compensating_txn BOOLEAN NOT NULL,
  amount_in_pennies BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT transactions_ledger_idempotency_key_check UNIQUE (idempotency_key, account_id, is_compensating_txn)
);

CREATE INDEX IF NOT EXISTS idx_transactions_ledger_account_id ON transactions_ledger(account_id);
