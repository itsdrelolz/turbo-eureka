package postgresdb

import (
	"context"
	"fmt"

	"job-matcher/internal/models"

	"github.com/google/uuid"
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

func (s *Store) Create(ctx context.Context, jobID uuid.UUID, fileName string) error {

	initalStatus := models.Queued

	sql := `
		INSERT into jobs (id, file_name, job_status)
		VALUES ($1, $2, $3)
		`

	err := s.Pool.QueryRow(
		ctx,
		sql,
		jobID,
		fileName,
		initalStatus.String(),
	)

	if err != nil {
		fmt.Errorf("ERROR: Query row failed with error: %w", err)
	}

	return nil

}

func (s *Store) ProcessingJob(ctx context.Context, jobID uuid.UUID) error {

}

func (s *Store) CompleteJob(ctx context.Context, jobID uuid.UUID, content string) error {

}
func (s *Store) Get(ctx context.Context, jobID uuid.UUID) (*models.Job, error) {

	var retrievedJob models.Job

	// convert to string before sending back
	var statusString string

	sql := `
        SELECT id, job_status, file_name, created_at
        FROM jobs
        WHERE id = $1
        `

	err := s.Pool.QueryRow(
		ctx,
		sql,
		jobID,
	).Scan(
		&retrievedJob.ID,
		&statusString,
		&retrievedJob.FileName,
		&retrievedJob.CreatedAt,
	)

	if err != nil {
		return &models.Job{}, fmt.Errorf("ERROR: Failed to retrieve job with error: %w", err)
	}

	jobStatus, err := models.StringToJobStatus(statusString)

	if err != nil {
		return nil, fmt.Errorf("ERROR: database contains invalid job status string: %w", err)
	}
	retrievedJob.JobStatus = jobStatus

	return &retrievedJob, nil

}

func (s *Store) GetResult(ctx context.Context, jobID uuid.UUID) (*models.Job, error) {

	var retrievedJob models.Job

	// convert to string before sending back
	var statusString string

	sql := `
        SELECT id, job_status, file_name, created_at, resume_content
        FROM jobs
        WHERE id = $1
        `

	err := s.Pool.QueryRow(
		ctx,
		sql,
		jobID,
	).Scan(
		&retrievedJob.ID,
		&statusString,
		&retrievedJob.FileName,
		&retrievedJob.CreatedAt,
		&retrievedJob.ResumeText,
	)

	if err != nil {
		return &models.Job{}, fmt.Errorf("ERROR: Failed to retrieve job with error: %w", err)
	}

	jobStatus, err := models.StringToJobStatus(statusString)

	if err != nil {
		return nil, fmt.Errorf("ERROR: database contains invalid job status string: %w", err)
	}
	retrievedJob.JobStatus = jobStatus

	return &retrievedJob, nil

}
