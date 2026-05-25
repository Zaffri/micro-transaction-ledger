package main

import (
	"context"
	_ "embed"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

// TODO: use proper migration tool: read all scripts from migration/*, and remove copies of seed insert SQL

//go:embed migrations/01_create_accounts_table.sql
var schemaSql string

//go:embed seeding/01_create_accounts.sql
var seedAccountSql string

//go:embed seeding/02_create_transaction.sql
var seedTransactionSql string

//go:embed seeding/03_create_ledger_entry.sql
var seedLedgerSql string

func main() {
	ctx := context.Background()
	log.Println("Starting migration...")
	dbPool, err := pgxpool.New(ctx, os.Getenv("DATABASE_CONNECTION"))

	if err != nil {
		log.Fatalf("Migration: Unable to connect to database: %v\n", err)
	}

	defer dbPool.Close()
	run(ctx, dbPool)
}

func run(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, schemaSql)

	if err != nil {
		log.Fatalf("Failed to run database migration: %v", err)
	}

	txn, err := db.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		log.Fatalf("Failed to begin seed txn: %v", err)
	}

	defer txn.Rollback(ctx)

	err = seedData(ctx, txn)

	if err != nil {
		log.Fatalf("Failed to seed txn: %v", err)
	}

	err = txn.Commit(ctx)

	if err != nil {
		log.Fatalf("Failed to commit seed txn: %v", err)
	}

	migrator, err := rivermigrate.New(riverpgxv5.New(db), nil)
	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)

	if err != nil {
		log.Fatalf("Failed to run River migrations: %v", err)
	}
	return err
}

func seedData(ctx context.Context, txn pgx.Tx) error {
	var userOneAccountId int64
	var userOneBalance int64
	err := txn.QueryRow(ctx, seedAccountSql, 1000, "John Doe").Scan(&userOneAccountId, &userOneBalance)

	if err != nil {
		return err
	}

	var userTwoAccountId int64
	var userTwoBalance int64
	err = txn.QueryRow(ctx, seedAccountSql, 5000, "Jane Doe").Scan(&userTwoAccountId, &userTwoBalance)

	if err != nil {
		return err
	}

	var transactionOneId int64
	var transactionOneStatus string
	err = txn.QueryRow(
		ctx,
		seedTransactionSql,
		"Transfer...test",
		userOneAccountId,
		"settled",
	).Scan(&transactionOneId, &transactionOneStatus)

	if err != nil {
		return err
	}

	idempotencyKey := "3b886308-9312-4178-a7fe-b77486732b77"

	// entry for account 1
	var ledgerEntryOneSenderId int64
	err = txn.QueryRow(
		ctx,
		seedLedgerSql,
		transactionOneId,
		userOneAccountId,
		userTwoAccountId,
		idempotencyKey,
		false,
		-300,
	).Scan(&ledgerEntryOneSenderId)

	if err != nil {
		return err
	}

	// entry for account 2
	var ledgerEntryOneReceiverId int64
	err = txn.QueryRow(
		ctx,
		seedLedgerSql,
		transactionOneId,
		userTwoAccountId,
		userOneAccountId,
		idempotencyKey,
		false,
		300,
	).Scan(&ledgerEntryOneReceiverId)

	if err != nil {
		return err
	}

	return nil
}
