package postgresdb

import (
	"context"
	"fmt"
	// pool required in order to handle concurrent access
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"job-matcher/internal/storage"
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

func (s *Store) InsertJobReturnID(ctx context.Context, fileUrl string, jobStatus storage.JobStatus) (uuid.NullUUID, error) {

	var newId uuid.NullUUID

	sql := `
		INSERT into jobs (file_url, job_status)
		VALUES ($1, $2)
		RETURNING id
		`

	err := s.Pool.QueryRow(
		ctx,
		sql,
		fileUrl,
		jobStatus.String(),
	).Scan(&newId)

	if err != nil {
		return uuid.NullUUID{}, fmt.Errorf("Query row failed with error: %w", err)
	}
	return newId, nil

}

func (s *Store) JobByID(ctx context.Context, jobID uuid.NullUUID) (storage.Job, error) {

	var retrievedJob storage.Job

	sql := `
        SELECT id, job_status, file_url, created_at
        FROM jobs
        WHERE id = $1
        `

	err := s.Pool.QueryRow(
		ctx,
		sql,
		jobID,
	).Scan(&retrievedJob.ID, &retrievedJob.JobStatus, &retrievedJob.FileUrl, &retrievedJob.CreatedAt)

	if err != nil {
		return storage.Job{}, fmt.Errorf("Failed to retrieve job with error: %w", err)
	}

	return retrievedJob, nil

}

func (s *Store) UpdateJobStatus(ctx context.Context, jobID uuid.NullUUID, jobStatus storage.JobStatus) error {

	var updatedJob storage.Job

	sql := `
	UPDATE job
	SET job_status = $1, 
	WHERE id = $2
	`

	err := s.Pool.QueryRow(
		ctx,
		sql,
		jobStatus.String(),
		jobID,
	).Scan(&updatedJob.ID, &updatedJob.JobStatus, &updatedJob.FileUrl, updatedJob.CreatedAt)

	if err != nil {
		return fmt.Errorf("Failed to update job with error: %w", err)
	}

	return nil

}

func (s *Store) String(js storage.JobStatus) string {
	switch js {
	case storage.Pending:
		return "pending"
	case storage.Queued:
		return "queued"
	case storage.Failed:
		return "failed"
	case storage.Completed:
		return "completed"
	default:
		return fmt.Sprintf("JobStatus(%d)", js)
	}
}
