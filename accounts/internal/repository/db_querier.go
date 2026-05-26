package repository

import "github.com/jackc/pgx/v5"

type DbQueries interface {
	Querier
	WithTx(tx pgx.Tx) *Queries
}
