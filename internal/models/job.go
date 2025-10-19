package pkg

import (
	"time"
)

type Status int

const (
	Queued Status = iota
	Done
)

type Job struct {
	ID        string
	FileKey   string
	Status    Status
	CreatedAt time.Time
}
