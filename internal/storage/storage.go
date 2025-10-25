package storage 


import ( 
	"context" 
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
	ID string 
	JobStatus JobStatus
	FileUrl string 
	CreatedAt time.Time 
}

type JobStore interface { 	
	InsertJobAndGetID(ctx context.Context, fileUrl string) (string, error) 
	GetJobByID(ctx context.Context, jobID string) (Job, error)
	String(js JobStatus) string 
}

