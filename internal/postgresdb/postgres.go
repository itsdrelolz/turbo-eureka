package postgresdb

import (
	"context"
	"errors"
	"fmt"

	// pool required in order to handle concurrent access
	"job-matcher/internal/storage"
	apperrors "job-matcher/internal/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
)

type Store struct {
	Pool *pgxpool.Pool
}


func New(ctx context.Context, connString string) (*Store, error) {
    if connString == "" {
        return nil, fmt.Errorf("database connection string is required")
    }

    // Configure connection to register pgvector types automatically
    config, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return nil, fmt.Errorf("failed to parse database config: %w", err)
    }

    // Register pgvector types on connect
    config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
        // Register pgvector data type with the connection
        if err := pgxvec.RegisterTypes(ctx, conn); err != nil {
            return fmt.Errorf("failed to register pgvector types: %w", err)
        }

        // Ensure the vector extension exists
        _, err := conn.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS vector`)
        if err != nil {
            return fmt.Errorf("failed to enable pgvector extension: %w", err)
        }

        return nil
    }

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

func (s *Store) InsertJobReturnID(ctx context.Context, fileUrl string, jobStatus storage.JobStatus) (uuid.UUID, error) {

	var newId uuid.UUID

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
		return uuid.Nil, fmt.Errorf("Query row failed with error: %w", err)
	}
	return newId, nil

}

func (s *Store) JobByID(ctx context.Context, jobID uuid.UUID) (*storage.Job, error) {

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

	var notFoundErr *pgconn.PgError

	if errors.As(err, &notFoundErr) {
		return nil, fmt.Errorf("Row not inserted in DB (404): %w", apperrors.ErrPermanentFailure)
	}

	if err != nil {
		return &storage.Job{}, fmt.Errorf("Failed to retrieve job with error: %w", err)
	}

	return &retrievedJob, nil

}

func (s *Store) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, jobStatus storage.JobStatus) error {

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

func (s *Store) SetEmbeddingWithID(ctx context.Context, jobID uuid.UUID, resumeEmbedding []float32) error {

	sql := `
		UPDATE jobs
		SET vector = $1
		WHERE id = $2
		`

	_, err := s.Pool.Exec(
		ctx,
		sql,
		pgvector.NewVector(resumeEmbedding),
		jobID,
	)

	var notFoundErr *pgconn.PgError

	if errors.As(err, &notFoundErr) {
		return fmt.Errorf("Embedding row not found (404): %w", apperrors.ErrPermanentFailure)
	}

	if err != nil {
		return fmt.Errorf("Failed to insert embedding into job with error: %w", err)
	}

	return nil

}

func (s *Store) String(js storage.JobStatus) string {
	switch js {
	case storage.Processing:
		return "processing"
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
