package processor

import (
	"context"
	"job-matcher/internal/gemini"
	"job-matcher/internal/objectstore"
	"job-matcher/internal/storage"
	"log"
	"fmt"
	"log/slog"
	"os"
	"time"
	"github.com/google/uuid"
)

type JobFetcher interface {
	ConsumeJob(ctx context.Context) (string, error)
}


type JobProcessor struct {
	db       storage.JobStore
	queue    JobFetcher
	store    objectstore.FileStorer
	s3Bucket string
	gemini   gemini.ResumerParser
}

func NewJobProcessor(db storage.JobStore, queue JobFetcher, store objectstore.FileStorer, s3Bucket string) *JobProcessor {
	return &JobProcessor{db: db, queue: queue, store: store, s3Bucket: s3Bucket}
}


/*
ADD failure recover for jobs as well. A job should know when to quit and try another job.
*/

func (p *JobProcessor) Run(ctx context.Context) {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)
	
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


	job, err := p.handleJobWithRetry(ctx, jobID)

	

}

func (p *JobProcessor) handleJobWithRetry(ctx context.Context, jobID uuid.UUID) (*storage.Job, error) { 

	const maxRetries = 3 


	for i := 0; i < maxRetries; i++ { 

		job, err := p.db.JobByID(ctx, jobID)

		if err == nil || !isRetryable(err) { 
			return job, err 
		}

	log.Printf("Retrying fetch for job %s (attempt %d): %v", jobID, i + 1, err) 

	// Exponential backoff
	// multiplicatively decrease the rate of some process, in order to gradually find an acceptable rate
	time.Sleep(time.Duration(1 << (i) * time.Second))
	}
	return nil, fmt.Errorf("failed to fetch job %s after %d retries", jobID, maxRetries)
}



// Function to process the job files 
// Updates the job status to processing, downloads the resume into memory. 
// Uses the gemini api to extract the text. 
// Then passes the text and uses gemini to create an embedding, 
// Returns the extracted text and resulting embedding 

func (p *JobProcessor) processJobFile(ctx context.Context, jobID uuid.UUID, job *storage.Job) (string, []float32, error){ 


	if err := p.db.UpdateJobStatus(ctx, jobID, storage.Processing); err != nil { 
		fmt.Errorf("Failed to update job status for job %s, with %v", jobID, err)
		return "", nil, err
	}


	userResume, err := p.store.Download(ctx, p.s3Bucket, job.FileUrl)

	if err != nil {
        log.Fatalf("Failed downloading file for job %s: %v", jobID, err)
        return "", nil, err 
    
	}


	extracedText, err := p.gemini.ExtractText(ctx, userResume)

	if err != nil {
        log.Fatalf("Failed to extract resume text for job %s: %v", jobID, err)
        return "", nil, err 
    }


	embeddedResume, err := p.gemini.Embed(ctx, extracedText)

	if err != nil { 
        log.Fatalf("Failed embedding for job %s: %v", jobID, err)
        return "", nil, err
    }

	return extracedText, embeddedResume, nil



}


// function is the final step to processing a job
// Its purpose is to save the progress made even if the process manages to fail on insertion 
func (p *JobProcessor) saveResultsWithRetry(ctx context.Context, jobID uuid.UUID, embeddings []float32) error { 

	err := p.db.SetEmbeddingWithID(ctx, jobID, embeddings)

	if err != nil { 
		return fmt.Errorf("Failed to update job status for job: %s, with: %v", jobID, err)
	}
	return nil

}

func isRetryable(err error) bool {
	return false
}
