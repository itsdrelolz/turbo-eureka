package database

import (
	"context"
	"fmt"

	// pool required in order to handle concurrent access 
	"github.com/jackc/pgx/v5/pgxpool"
)


type Store struct { 
	Pool *pgxpool.Pool
}


func New(ctx context.Context, connString string) (*Store, error) { 
	
	if connString == "" { 
		return nil, fmt.Errorf("database connection string is required")	
	}

	pool, err := pgxpool.New(ctx, connString)

	if err != nil { 
		return nil, fmt.Errorf("unable to create connection: %w", err)
	}

	if err := pool.Ping(ctx); err != nil { 
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}


	return &Store{Pool: pool}, nil


}


func (s *Store) Close() { 
	s.Pool.Close() 
}


