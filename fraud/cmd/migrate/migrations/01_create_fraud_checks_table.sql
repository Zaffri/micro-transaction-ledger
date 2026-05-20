CREATE EXTENSION IF NOT EXISTS moddatetime;

CREATE TABLE IF NOT EXISTS fraud_checks (
  id BIGSERIAL PRIMARY KEY,
  transaction_id BIGINT NOT NULL,
  risk_score smallint NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS trigger_update_fraud_checks_timestamp ON fraud_checks;

CREATE TRIGGER trigger_update_fraud_checks_timestamp
  BEFORE UPDATE ON fraud_checks
  FOR EACH ROW
  EXECUTE PROCEDURE moddatetime(updated_at);
