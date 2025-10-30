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
	ID        uuid.NullUUID
	JobStatus JobStatus
	FileUrl   string
	CreatedAt time.Time
}

type JobStore interface {
	InsertJobReturnID(ctx context.Context, fileUrl string, jobStatus JobStatus) (uuid.NullUUID, error)
	JobByID(ctx context.Context, jobID uuid.NullUUID) (Job, error)
	UpdateJobStatus(ctx context.Context, jobID uuid.NullUUID, jobStatus JobStatus) error
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
