package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewDB(cfg Config) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("db connect error: %w", err)
	}
	return conn, nil
}
