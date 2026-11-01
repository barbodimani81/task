// Package greetings reads greeting messages from Postgres.
package greetings

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound means there is no greeting for the requested language.
var ErrNotFound = errors.New("greeting not found")

// Store reads greetings from the greetings table.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore returns a Store that queries through pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Message returns the greeting for language, or ErrNotFound.
func (s *Store) Message(ctx context.Context, language string) (string, error) {
	var msg string
	// $1 keeps the value out of the SQL text, so it can't be used for injection.
	err := s.pool.QueryRow(ctx,
		"SELECT message FROM greetings WHERE language = $1", language,
	).Scan(&msg)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return msg, err
}

// Ping reports whether the database is reachable.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
