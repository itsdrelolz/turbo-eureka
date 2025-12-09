package models

import (
	"time"

	"github.com/google/uuid"
)

type Status string 

const (
	StatusQueued     Status = "QUEUED"
	StatusProcessing Status = "PROCESSING"
	StatusCompleted  Status = "COMPLETED"
	StatusFailed     Status = "FAILED"
)

type Job struct {
	ID uuid.UUID `json:"id" db:"id"`

	Status string `json:"status" db:"status"`

	FileName string `json:"file_name" db:"file_name"`

	ResumeText *string `json:"resume_text,omitempty" db:"resume_text"`
 
	ErrorMessage *string `json:"error_message,omitempty" db:"error_message"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`

}
