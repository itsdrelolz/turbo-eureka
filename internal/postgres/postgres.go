package postgres

import (
	"context"
	"errors"
	"fmt"
	"job-matcher/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (s *Store) Create(ctx context.Context, job *models.Job) error {

	sql := `
		INSERT into jobs (id, file_name)
		VALUES ($1, $2)
		`

	_, err := s.Pool.Exec(
		ctx,
		sql,
		job.ID,
		job.FileName,
	)
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, jobID uuid.UUID) (*models.Job, error) {

	sql := `
        SELECT *
        FROM jobs
        WHERE id = $1
    `

	rows, err := s.Pool.Query(ctx, sql, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to query job: %w", err)
	}

	job, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[models.Job])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("job not found: %w", err)
		}
		return nil, fmt.Errorf("failed to retrieve job with error: %w", err)
	}

	return job, nil
}

func (s *Store) CompleteJob(ctx context.Context, jobID uuid.UUID, data string) error {
	sql := `UPDATE jobs 
              SET status = $1, resume_text = $2
              WHERE id = $3`

	commandTag, err := s.Pool.Exec(
		ctx,
		sql,
		models.StatusCompleted,
		data,
		jobID,
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %v", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no job found with id %s", jobID)
	}

	return nil

}

func (s *Store) FailJob(ctx context.Context, jobID uuid.UUID, errMsg error) error {
	sql := `UPDATE jobs 
              SET status = $1, error_message = $2 
              WHERE id = $3`

	var errString string

	if errMsg != nil {
		errString = errMsg.Error()
	}
	commandTag, err := s.Pool.Exec(
		ctx,
		sql,
		models.StatusFailed,
		errString,
		jobID,
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %v", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no job found with id %s", jobID)
	}
	return nil

}
