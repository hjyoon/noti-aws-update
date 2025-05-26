package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func NewDB(cfg Config) (*pgx.Conn, error) {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		return nil, fmt.Errorf("db connect error: %w", err)
	}
	return conn, nil
}
