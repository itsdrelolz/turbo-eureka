package queue

import (
	"context"
)

type JobProducer interface {
	InsertJob(ctx context.Context, jobID string) error
}

type JobConsumer interface {
	ConsumeJob(ctx context.Context) (string, error)
}

type JobQueue interface {
	JobProducer
	JobConsumer
}
