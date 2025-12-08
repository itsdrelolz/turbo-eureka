package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID uuid.UUID

	JobStatus JobStatus

	FileName string

	CreatedAt time.Time

	ResumeText string
}
