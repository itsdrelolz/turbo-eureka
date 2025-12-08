package postgresdb

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"job-matcher/internal/models"
)

type Store struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*Store, error) {
	if connString == "" {
		return nil, fmt.Errorf("ERROR: database connection string is required")
	}

	config, err := pgxpool.ParseConfig(connString)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("ERROR: unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ERROR: unable to ping database: %w", err)
	}

	return &Store{Pool: pool}, nil
}

func (s *Store) Close() {
	s.Pool.Close()
}

func (s *Store) Create(ctx context.Context, jobID uuid.UUID, fileName string) {

	sql := `
		INSERT into jobs (id, file_name)
		VALUES ($1, $2, $3)
		`

	s.Pool.QueryRow(
		ctx,
		sql,
		jobID,
		fileName,
	)

}

func (s *Store) Get(ctx context.Context, jobID uuid.UUID) (*models.Job, error) {

	var retrievedJob models.Job

	sql := `
        SELECT id, file_name
        FROM jobs
        WHERE id = $1
        `

	err := s.Pool.QueryRow(
		ctx,
		sql,
		jobID,
	).Scan(
		&retrievedJob.ID,
		&retrievedJob.FileName,
	)

	if err != nil {
		return &models.Job{}, fmt.Errorf("ERROR: Failed to retrieve job with error: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("ERROR: database contains invalid job status string: %w", err)
	}

	return &retrievedJob, nil
}


func (s *Store) CompleteJob(ctx context.Context, jobID uuid.UUID, data string) error {
    sql := `UPDATE jobs 
              SET status = 'COMPLETED', result_text = $1, updated_at = NOW() 
              WHERE id = $2`

	err := s.Pool.QueryRow(
		ctx,
		sql,
		jobID,
	)

	if err != nil { 
		return fmt.Errorf(err) 
	}

	return nil

}

func (s *Store) FailedJob(ctx context.Context, jobID uuid.UUID, errMsg error) error {
    sql := `UPDATE jobs 
              SET status = 'FAILED', error_message = $1, updated_at = NOW() 
              WHERE id = $2`
	err := s.Pool.QueryRow(
		ctx,
		sql,
		errMsg.Error(),
		jobID,
	)

	if err != nil { 
		return fmt.Errorf(err) 
	}

	return nil


}
