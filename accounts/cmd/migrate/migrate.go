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
var seedSql string

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

	_, err = txn.Exec(ctx, seedSql, 1000, "John Doe")

	if err != nil {
		log.Fatalf("Failed to seed txn: %v", err)
	}

	_, err = txn.Exec(ctx, seedSql, 5000, "Jane Doe")

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
