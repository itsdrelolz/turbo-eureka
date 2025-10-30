package storage

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type JobStatus int

const (
	Pending JobStatus = iota
	Queued
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
	JobByID(ctx context.Context, jobID uuid.UUID) (Job, error)
	UpdateJobStatus(ctx context.Context, jobID uuid.UUID, jobStatus JobStatus) error
	InsertEmbeddingWithID(ctx context.Context, jobID uuid.UUID, resumeEmbedding []byte) error 
}

func (s JobStatus) String() string {
	switch s {
	case Queued:
		return "queued"
	case Pending:
		return "pending"
	case Completed:
		return "completed"
	case Failed:
		return "failed"
	default:
		return "unknown"
	}
}
