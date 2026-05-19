package main

import (
	"context"
	_ "embed"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

// TODO: use proper migration tool: read all scripts from migration/*, and remove copies of seed insert SQL

//go:embed migrations/01_create_accounts_table.sql
var schemaSql string

//go:embed seeding/01_create_accounts.sql
var seedSql string

func main() {
	ctx := context.Background()
	log.Printf("Connecting.... %s", os.Getenv("DATABASE_CONNECTION"))
	db, err := pgx.Connect(ctx, os.Getenv("DATABASE_CONNECTION"))

	if err != nil {
		log.Printf("Migration: Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	defer db.Close(ctx)
	run(ctx, db)
}

func run(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, schemaSql)

	if err != nil {
		log.Printf("Failed to run database migration: %v", err)
		return err
	}

	txn, err := db.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		log.Printf("Failed to begin seed txn: %v", err)
		return err
	}

	defer txn.Rollback(ctx)

	_, err = txn.Exec(ctx, seedSql, 1000, "John Doe")

	if err != nil {
		log.Printf("Failed to seed txn: %v", err)
		return err
	}

	_, err = txn.Exec(ctx, seedSql, 5000, "Jane Doe")

	if err != nil {
		log.Printf("Failed to seed txn: %v", err)
		return err
	}

	err = txn.Commit(ctx)

	if err != nil {
		log.Printf("Failed to commit seed txn: %v", err)
		return err
	}

	return err
}
