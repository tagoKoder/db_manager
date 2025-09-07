package db_manager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// CanConnectDSN tries to connect using the DSN as-is.
// Returns true if it connects successfully. If it fails, the error indicates the cause.
// If the database in the DSN does not exist, it tries to report it clearly.
func CanConnectDSN(dsn string) (bool, error) {
	timeout := 5 * time.Second
	// NOTE: "postgres" = lib/pq; if you use pgx stdlib it would be "pgx".
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return false, fmt.Errorf("while opening the driver: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		// Explicit timeout?
		if errors.Is(err, context.DeadlineExceeded) {
			return false, fmt.Errorf("connection timeout (%s): %w", timeout, err)
		}

		// Detect "database does not exist" by code 3D000 (pgx)
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) && pgxErr.Code == "3D000" {
			return false, fmt.Errorf("the database specified in the DSN does not exist (code 3D000): %w", err)
		}
		// Detect the same in lib/pq
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == "3D000" {
			return false, fmt.Errorf("the database specified in the DSN does not exist (code 3D000): %w", err)
		}

		// Any other cause (credentials, network, SSL, etc.)
		return false, fmt.Errorf("could not connect with the DSN: %w", err)
	}

	return true, nil
}
