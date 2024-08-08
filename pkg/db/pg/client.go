package pg

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	"github.com/mikhailsoldatkin/platform_common/pkg/db"
)

// pgClient implements the db.Client interface for PostgreSQL using pgxpool.
type pgClient struct {
	masterDBC db.DB
}

// New creates a new instance of pgClient and connects to the PostgreSQL database.
func New(ctx context.Context, dsn string) (db.Client, error) {
	dbc, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, errors.Errorf("failed to connect to db: %v", err)
	}

	return &pgClient{masterDBC: &pg{dbc: dbc}}, nil
}

// DB returns the database client instance associated with pgClient.
func (c *pgClient) DB() db.DB {
	return c.masterDBC
}

// Close closes the database connection. It returns an error if closing the connection fails.
func (c *pgClient) Close() error {
	if c.masterDBC != nil {
		c.masterDBC.Close()
	}

	return nil
}
