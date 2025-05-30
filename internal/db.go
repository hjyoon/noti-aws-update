package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDBPool(cfg Config) (*pgxpool.Pool, error) {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	var (
		pool        *pgxpool.Pool
		err         error
		maxAttempts = 5
	)

	for i := 1; i <= maxAttempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pool, err = pgxpool.New(ctx, dbUrl)
		if err == nil {
			pingErr := pool.Ping(ctx)
			if pingErr == nil {
				return pool, nil
			} else {
				err = pingErr
			}
		}

		fmt.Printf("DB connection attempt %d/%d failed: %v\n", i, maxAttempts, err)
		if i < maxAttempts {
			time.Sleep(time.Duration(i) * time.Second)
		}
	}

	return nil, fmt.Errorf("db connect failed after %d attempts: %w", maxAttempts, err)
}
