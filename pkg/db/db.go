package db

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// Handler is a function that is executed within a transaction.
type Handler func(ctx context.Context) error

// Client represents a client for working with the database.
type Client interface {
	DB() DB
	Close() error
}

// TxManager is a transaction manager that executes the specified handler within a transaction.
type TxManager interface {
	ReadCommitted(ctx context.Context, f Handler) error
}

// Query is a wrapper around a query, containing the query name and the query itself.
// The query name is used for logging and potentially could be used for tracing or other purposes.
type Query struct {
	Name     string
	QueryRaw string
}

// Transactor is an interface for working with transactions.
type Transactor interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// SQLExecer combines NamedExecer and QueryExecer.
type SQLExecer interface {
	NamedExecer
	QueryExecer
}

// NamedExecer is an interface for working with named queries using tags in structures.
type NamedExecer interface {
	ScanOneContext(ctx context.Context, dest any, q Query, args ...any) error
	ScanAllContext(ctx context.Context, dest any, q Query, args ...any) error
}

// QueryExecer is an interface for working with regular queries.
type QueryExecer interface {
	ExecContext(ctx context.Context, q Query, args ...any) (pgconn.CommandTag, error)
	QueryContext(ctx context.Context, q Query, args ...any) (pgx.Rows, error)
	QueryRowContext(ctx context.Context, q Query, args ...any) pgx.Row
}

// Pinger is an interface for checking the connection to the database.
type Pinger interface {
	Ping(ctx context.Context) error
}

// DB is an interface for working with the database.
type DB interface {
	SQLExecer
	Transactor
	Pinger
	Close()
}
