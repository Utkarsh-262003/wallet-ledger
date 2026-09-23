package main

import "github.com/jackc/pgx/v5/pgxpool"

// Store holds all the SQL the wallet service runs.
type Store struct {
	pool *pgxpool.Pool
}