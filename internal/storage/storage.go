package storage

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type JobStatus int

const (
	Queued JobStatus = iota
	Processing
	Failed
	Completed
)

type Job struct {
	ID        uuid.UUID
	JobStatus JobStatus
	FileUrl   string
	CreatedAt time.Time
}

type JobStore interface {
	InsertJobReturnID(ctx context.Context, fileUrl string, jobStatus JobStatus) (uuid.UUID, error)
	JobByID(ctx context.Context, jobID uuid.UUID) (*Job, error)
	UpdateJobStatus(ctx context.Context, jobID uuid.UUID, jobStatus JobStatus) error
	SetEmbeddingWithID(ctx context.Context, jobID uuid.UUID, resumeEmbedding []float32) error
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
