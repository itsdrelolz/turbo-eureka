package queue 



import (
	"context" 
)


type JobQueuer interface { 	
	InsertJob(ctx context.Context,  jobID string) error 
}
