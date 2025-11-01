package processor

import (
	"context"
	"fmt"
	"job-matcher/internal/gemini"
	"job-matcher/internal/objectstore"
	"job-matcher/internal/storage"
	"log"
	"log/slog"
	"math"
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

// logic for handling single job

func (p *JobProcessor) processJob(ctx context.Context, jobID uuid.UUID) {

	log.Printf("Processing given job for ID: %s", jobID)


	const totalJobTimeout = 30 * time.Second
	// if this fails, mark the job as failed

	jobCtx, cancel := context.WithTimeout(ctx, totalJobTimeout)

	
	job, err := p.fetchJobWithRetry(jobCtx, jobID)

	if err != nil {

		log.Printf("Faild job with error: %v", err)
	}

	resumeEmbedding, err := p.processJobFile(jobCtx, job, job.ID)

	if err != nil { 
		log.Printf("Faild job with error: %w", err)
	}

	err = p.saveResultsWithRetry(jobCtx, job.ID, resumeEmbedding)

	cancel()


}


func (p *JobProcessor) fetchJobWithRetry(ctx context.Context, jobID uuid.UUID)  (*storage.Job, error) { 

	const maxRetries = 3 

	// database calls can be more frequent
	const baseDelay = 1 * time.Second

	for i := range maxRetries { 

		job, err := p.db.JobByID(ctx, jobID)

		if err == nil { 
			return job, nil
		}

		if !isRetryable(err) { 

			return nil, fmt.Errorf("non-retryable error fetching job %s: %w", jobID, err)
		}
// 3. Transient Failure: Log, Sleep, and Retry
        slog.Warn("Transient error fetching job, retrying...", 
            "jobID", jobID, "attempt", i+1, "error", err)
            
        // Use backoff only between attempts (not on the last attempt)
        if i < maxRetries-1 {
            delay := calculateExponentialBackoff(i, baseDelay)
            time.Sleep(delay)
        }
    }
    
    // 4. Final Failure: Max attempts reached
    return nil, fmt.Errorf("failed to fetch job %s after %d retries: %w", jobID, maxRetries, ctx.Err())
}




// Function to process the job files 
// Updates the job status to processing, downloads the resume into memory. 
// Uses the gemini api to extract the text. 
// Then passes the text and uses gemini to create an embedding, 
// Returns the extracted text and resulting embedding 

func (p *JobProcessor) processJobFile(ctx context.Context, job *storage.Job, jobID uuid.UUID) ([]float32, error){ 


	if err := p.updateJobWithRetry(ctx, jobID, storage.Processing); err != nil { 
		fmt.Errorf("Failed to update job status for job %s, with %v", jobID, err)
		return nil, err
	}


	userResume, err := p.store.Download(ctx, p.s3Bucket, job.FileUrl)

	if err != nil {
        log.Printf("Failed downloading file for job %s: %v", jobID, err)
        return nil, err 
    
	}


	extracedText, err := p.gemini.ExtractText(ctx, userResume)

	if err != nil {
        log.Printf("Failed to extract resume text for job %s: %v", jobID, err)
        return nil, err 
    }


	embeddedResume, err := p.gemini.Embed(ctx, extracedText)

	if err != nil { 
        log.Printf("Failed embedding for job %s: %v", jobID, err)
        return nil, err
    }

	return embeddedResume, nil
}


func (p *JobProcessor) updateJobWithRetry(ctx context.Context, jobID uuid.UUID, jobStatus storage.JobStatus) error { 


	const maxRetries = 3
	const baseDelay = 10

	for i := range maxRetries { 

	
		err := p.db.UpdateJobStatus(ctx, jobID, jobStatus)

		if err == nil || !isRetryable(err) { 
			return err
		}	

		log.Printf("Retrying failed job %s, (attempt %d). With error: %v", jobID, i + 1, err)

		

	time.Sleep(time.Duration(calculateExponentialBackoff(i, baseDelay)))

	}


	return fmt.Errorf("failed to fetch job %s after %d retries", jobID, maxRetries)
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



// calulates the exponential backoff by baseDelay * 2^attempts
func calculateExponentialBackoff(attempts int, baseDelaySeconds time.Duration) time.Duration { 
	backoff := float64(baseDelaySeconds) * math.Pow(2, float64(attempts))

	return time.Duration(backoff) * time.Second
}
