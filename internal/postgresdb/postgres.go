package postgresdb

import (
	"context"
	"fmt"
	// pool required in order to handle concurrent access
	"github.com/jackc/pgx/v5/pgxpool"
	"job-matcher/internal/storage"
	"github.com/google/uuid"
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

func (s *Store) InsertJobAndGetID(ctx context.Context, fileUrl string) (uuid.NullUUID, error) {
	
	var newId uuid.NullUUID
	jobStatus := storage.Queued

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


func (s *Store) GetJobByID(ctx context.Context, jobID uuid.NullUUID) (storage.Job, error) { 
	
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
