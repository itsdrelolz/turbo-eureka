package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Status int

// default status in database is set as value 1 or "QUEUED" 
const ( 
	StatusUnknown Status = iota 
	StatusQueued           // 1 
	StatusPending          // 2 
	StatusCompleted        // 3 
	StatusFailed           // 4
)

type Job struct {
	ID uuid.UUID `json:"id" db:"id"`

	Status Status `json:"status" db:"status"`

	FileName string `json:"file_name" db:"file_name"`

	ResumeText *string `json:"resume_text,omitempty" db:"resume_text"`
 
	ErrorMessage *string `json:"error_message,omitempty" db:"error_message"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`

}


func (s Status) String() string { 
	switch s { 
	case StatusQueued:
		return "queued" 
	case StatusPending: 
		return "pending"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

func (s Status) MarshallJSON() ([]byte, error) { 
	return json.Marshal(s.String())
}


