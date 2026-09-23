package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store holds all the SQL the wallet service runs.
type Store struct {
	pool *pgxpool.Pool
}

var ErrNotFound = errors.New("not found")

type Wallet struct {
	ID           string
	UserID       string
	BalanceCents int64
	Currency     string
	Version      int32
	CreatedAt    time.Time
}

// The ::text casts turn UUIDs into plain strings as Postgres returns them.
const walletColumns = `id::text, user_id::text, balance_cents, currency, version, created_at`

func scanWallet(row pgx.Row) (Wallet, error) {
	var w Wallet
	err := row.Scan(&w.ID, &w.UserID, &w.BalanceCents, &w.Currency, &w.Version, &w.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Wallet{}, ErrNotFound
	}
	return w, err
}

// CreateWallet creates the user's wallet, or returns the existing one.
// Calling it twice for the same user is safe.
func (s *Store) CreateWallet(ctx context.Context, userID string) (Wallet, error) {
	// ON CONFLICT DO NOTHING would return no row when the wallet already exists.
	// A do-nothing UPDATE makes RETURNING hand back the existing row instead.
	return scanWallet(s.pool.QueryRow(ctx,
		`INSERT INTO wallet.wallets (user_id) VALUES ($1)
		 ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
		 RETURNING `+walletColumns, userID))
}

func (s *Store) GetWalletByUser(ctx context.Context, userID string) (Wallet, error) {
	return scanWallet(s.pool.QueryRow(ctx,
		`SELECT `+walletColumns+` FROM wallet.wallets WHERE user_id = $1`, userID))
}