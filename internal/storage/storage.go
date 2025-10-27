package storage 


import ( 
	"context" 
	"time"
	"github.com/google/uuid"
) 

type JobStatus int


const (
	Pending JobStatus = iota 
	Queued             
	Failed             
	Completed          
)


type Job struct { 	
	ID uuid.NullUUID 
	JobStatus JobStatus
	FileUrl string 
	CreatedAt time.Time 
}

type JobStore interface { 	
	InsertJobAndGetID(ctx context.Context, fileUrl string) (uuid.NullUUID, error) 
	GetJobByID(ctx context.Context, jobID uuid.NullUUID) (Job, error)
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
