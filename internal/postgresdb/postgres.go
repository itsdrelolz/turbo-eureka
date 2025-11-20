package postgresdb

import (
	"context"
	"errors"
	"fmt"
	"time"
	// pool required in order to handle concurrent access
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "job-matcher/internal/errors"
)

type JobStatus int

const (
	Queued JobStatus = iota

	Processing

	Failed

	Completed
)

type Job struct {
	ID uuid.UUID

	JobStatus JobStatus

	FileName string

	CreatedAt time.Time
}

type Store struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*Store, error) {
	if connString == "" {
		return nil, fmt.Errorf("database connection string is required")
	}

	config, err := pgxpool.ParseConfig(connString)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Store{Pool: pool}, nil
}

func (s *Store) Close() {
	s.Pool.Close()
}

func (s *Store) InsertJobReturnID(ctx context.Context, fileName string, jobStatus JobStatus) (uuid.UUID, error) {

	var newId uuid.UUID

	sql := `
		INSERT into jobs (file_name, job_status)
		VALUES ($1, $2)
		RETURNING id
		`

	err := s.Pool.QueryRow(
		ctx,
		sql,
		fileName,
		jobStatus.String(),
	).Scan(&newId)

	if err != nil {
		return uuid.Nil, fmt.Errorf("Query row failed with error: %w", err)
	}
	return newId, nil

}

func (s *Store) JobByID(ctx context.Context, jobID uuid.UUID) (*Job, error) {

	var retrievedJob Job

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

	var notFoundErr *pgconn.PgError

	if errors.As(err, &notFoundErr) {
		return nil, fmt.Errorf("Row not inserted in DB (404): %w", apperrors.ErrPermanentFailure)
	}

	if err != nil {
		return &Job{}, fmt.Errorf("Failed to retrieve job with error: %w", err)
	}

	jobStatus, err := stringToJobStatus(statusString)

	if err != nil {
		return nil, fmt.Errorf("database contains invalid job status string: %w", err)
	}
	retrievedJob.JobStatus = jobStatus

	return &retrievedJob, nil

}

func (s *Store) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, jobStatus JobStatus) error {

	sql := `
	UPDATE jobs
	SET job_status = $1
	WHERE id = $2
	`

	_, err := s.Pool.Exec(
		ctx,
		sql,
		jobStatus.String(),
		jobID,
	)

	var notFoundErr *pgconn.PgError

	if errors.As(err, &notFoundErr) {
		return fmt.Errorf("Job not found (404): %w", apperrors.ErrPermanentFailure)
	}

	if err != nil {
		return fmt.Errorf("Failed to update job with error: %w", err)
	}

	return nil

}

func (s *Store) SetContentWithID(ctx context.Context, jobID uuid.UUID, resumeContent string) error {

	sql := `
		UPDATE jobs
		SET resume_content = $1
		WHERE id = $2
		`

	_, err := s.Pool.Exec(
		ctx,
		sql,
		resumeContent,
		jobID,
	)

	var notFoundErr *pgconn.PgError

	if errors.As(err, &notFoundErr) {
		return fmt.Errorf("row not found (404): %w", apperrors.ErrPermanentFailure)
	}

	if err != nil {
		return fmt.Errorf("Failed to insert content into job with error: %w", err)
	}

	return nil

}

func stringToJobStatus(s string) (JobStatus, error) {
	switch s {
	case "queued":
		return Queued, nil
	case "processing":
		return Processing, nil
	case "completed":
		return Completed, nil
	case "failed":
		return Failed, nil
	default:
		// Return a default and an error
		return Failed, fmt.Errorf("unknown job status string: %s", s)
	}
}

func (s JobStatus) String() string {

	switch s {

	case Queued:

		return "queued"

	case Processing:

		return "processing"

	case Completed:

		return "completed"

	case Failed:

		return "failed"

	default:

		return "unknown"

	}

}
