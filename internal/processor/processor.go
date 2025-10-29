package processor

import (
	"context"
	"job-matcher/internal/objectstore"
	"job-matcher/internal/storage"
	"log"
	"time"

	"github.com/google/uuid"
)

type JobConsumer interface { 
	ConsumeJob(ctx context.Context) (string, error)
	UpdateJobStatus(ctx context.Context, jobID uuid.NullUUID) (uuid.NullUUID, error)
}

type JobUpdated interface { 
	UpdateJobStatus(ctx context.Context, jobID uuid.NullUUID) (uuid.NullUUID, error)
}
	


type JobProcessor struct { 

	db storage.JobStore 
	queue JobConsumer 
	store objectstore.FileStorer
	s3Bucket string
}


func NewJobProcessor(db storage.JobStore, queue JobConsumer, store objectstore.FileStorer, s3Bucket string) *JobProcessor { 
	return &JobProcessor{db: db, queue: queue, store: store, s3Bucket: s3Bucket}
}


func (p *JobProcessor) Run(ctx context.Context) { 
	log.Printf("Job processor has started, waiting for jobs...")

	// first thing is to now pull a job off of the worker queue 

	for { 

	jobIdStr, err := p.queue.ConsumeJob(ctx)


	if err != nil { 
		log.Printf("error consuming job from queue: %v", err)
		time.Sleep(5 * time.Second) // wait and then try again
		continue
	}

	jobID, err := uuid.Parse(jobIdStr)

	if err != nil { 
			log.Printf("invalid job id given, trying another job: %s", jobIdStr)
			continue
	}

	p.processJob(ctx, jobID)

	
	}

}







func (p *JobProcessor) processJob(ctx context.Context, jobID uuid.UUID) { 

	log.Printf("Processing given job for ID: %s", jobID)



	// got job details here, now its time to work on the details. This is extracting the pdfs text. 

	job, err := p.db.GetJobByID(ctx, uuid.NullUUID{UUID: jobID, Valid: true})

	if err != nil { 
		log.Printf("Error fetching job %s details: %v", jobID, err)
		return
	}

	// first update the job to have the status processing 
	// then it should start extracting the text using text extracting library 

	// UPDATED JOB STATUS 

	 err = p.db.UpdateJobStatus(ctx, uuid.NullUUID{UUID: jobID, Valid: true}, storage.Pending)
	
	if err != nil { 
		log.Printf("Failed updating status for job: %s details: %v", jobID, err)
		return 
	}

	// create download method to handle pdf
	// file is at most 5MB so its okay to pull it all in memory and pass it all to the library 







}
