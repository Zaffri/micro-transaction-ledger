INSERT INTO accounts (balance_in_pennies, account_holder_name)
VALUES ($1, $2)
RETURNING id, balance_in_pennies;
