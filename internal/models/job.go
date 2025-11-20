package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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

func StringToJobStatus(s string) (JobStatus, error) {
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
