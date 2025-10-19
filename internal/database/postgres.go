package database

import (
	"context"
	"fmt"
	// pool required in order to handle concurrent access
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobStatus int


const (
	Pending JobStatus = iota 
	Queued             
	Failed             
	Completed          
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

func (s *Store) InsertJobAndGetID(ctx context.Context, fileUrl string) (string, error) {
	
	var newId string 
	jobStatus := Queued

	sql := `
		INSERT into jobs (file_url, job_status)
		VALUES ($1, $2)
		RETURNING id
		`

	err := s.Pool.QueryRow(
		ctx, 
		sql,
		fileUrl,
		jobStatus,
		).Scan(&newId)


	if err != nil {
		return "", fmt.Errorf("Query row failed with error: %w", err)
	}
	return newId, nil

}



func (js JobStatus) String() string {
	switch js {
	case Pending:
		return "pending"
	case Queued:
		return "queued"
	case Failed:
		return "failed"
	case Completed:
		return "completed"
	default:
		return fmt.Sprintf("JobStatus(%d)", js)
	}
}
