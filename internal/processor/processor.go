package processor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	apperrors "job-matcher/internal/errors"
	"job-matcher/internal/objectstore"
	"job-matcher/internal/queue"
	"job-matcher/internal/storage"
	"log"
	"log/slog"
	"math"
	"os"
	"rsc.io/pdf"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// Set number of workers

const numWorkers = 5

type JobProcessor struct {
	db       storage.JobUpdater
	queue    queue.JobConsumer
	store    objectstore.FileFetcher
	s3Bucket string
}

func NewJobProcessor(db storage.JobUpdater, queue queue.JobConsumer, store objectstore.FileFetcher, s3Bucket string) *JobProcessor {
	return &JobProcessor{db: db, queue: queue, store: store, s3Bucket: s3Bucket}
}

// Runner for a job. Completed jobs will be marked as complete in the database. Failed jobs are marked incomplete.
func (p *JobProcessor) startConsumer(ctx context.Context, jobsChan chan uuid.UUID) {

	defer close(jobsChan)

	log.Printf("Job processor has started, waiting for jobs...")

	for {

		jobIDStr, err := p.queue.ConsumeJob(ctx)

		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Printf("Job consumer stopping...")
				return // Exit the loop and goroutine
			}
			log.Printf("error consuming from job queue, trying again...")
			time.Sleep(5 * time.Second) // wait and then try again
			continue
		}

		jobID, err := uuid.Parse(jobIDStr)
		if err != nil {
			log.Printf("invalid job id given, skipping job: %s. Error: %v", jobIDStr, err)
			continue
		}

		select {
		case jobsChan <- jobID:
			// Job successfully sent to a worker
		case <-ctx.Done():
			// Shutdown signal received while trying to send
			log.Printf("Job consumer stopping, not sending new job")
			return
		}
	}
}

func (p *JobProcessor) startWorker(ctx context.Context, workerID int, jobsChan <-chan uuid.UUID) {

	log.Printf("Worker #%d started", workerID)

	for jobID := range jobsChan {

		p.processJob(ctx, jobID)

	}
	log.Printf("Worker #%d stopped.", workerID)
}

func (p *JobProcessor) Run(ctx context.Context) {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	log.Printf("Job processor has started, launching %d workers...", numWorkers)

	jobsChan := make(chan uuid.UUID)

	var wg sync.WaitGroup

	// separate go routine for consuming jobs.
	go p.startConsumer(ctx, jobsChan)

	// starts the set amount of workers
	for i := 1; i <= numWorkers; i++ {

		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()
			p.startWorker(ctx, workerID, jobsChan)
		}(i)

	}

	<-ctx.Done()
	log.Printf("Shutdown signal receieved. Waiting for workers to finish...")

	wg.Wait()
	log.Printf("All workers stopped. Job processor shut down gracefully.")
}

// Logic for handling a jobs processings. The job has a set limit for how long it may run before cancelling.
// A job attemps to fetch the necessary information from the db
// It then attempts to generate the emedding of the given resume
// The job then updates the db with the resume content
func (p *JobProcessor) processJob(ctx context.Context, jobID uuid.UUID) {

	log.Printf("Processing given job with ID: %s", jobID)

	// hard timelimit set for the completion of the job
	// updating a jobs status to failed or completed bypasses this limit
	const totalJobTimeout = 2 * time.Minute
	jobCtx, cancel := context.WithTimeout(ctx, totalJobTimeout)
	defer cancel()

	job, err := p.fetchJobWithRetry(jobCtx, jobID)

	if err != nil {
		log.Printf("Faild job with error: %v", err)
		// unrecoverable job, mark as failed
		if !isRetryable(err) {
			updateCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			p.db.UpdateJobStatus(updateCtx, jobID, storage.Failed)
		}
		return
	}

	extractedText, err := p.processJobFile(jobCtx, job, job.ID)

	if err != nil {
		if errors.Is(err, apperrors.ErrPermanentFailure) {
			log.Printf("Fatal, non-retryable failure for job %s: %v. Marking as failed.", jobID, err)
			updateCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			p.db.UpdateJobStatus(updateCtx, jobID, storage.Failed)
			return // End job execution instantly
		}

		log.Printf("Unrecoverable error processing job %s: %v.", jobID, err)
		return
	}

	if err := p.saveResultsWithRetry(jobCtx, job.ID, extractedText); err != nil {
		log.Printf("Failed to save results for job %s: %v. Job failed.", jobID, err)
		if !isRetryable(err) {
			updateCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			p.db.UpdateJobStatus(updateCtx, jobID, storage.Failed)
		}
		return
	}

	// use a new, short-lived context to ensure this update succeeds
	// even if the jobCtx timed out
	updateCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if updateErr := p.updateJobWithRetry(updateCtx, jobID, storage.Completed); updateErr != nil {
		log.Printf("WARNING: Job %s completed work, but failed to mark status as Completed: %v", jobID, updateErr)
	} else {
		log.Printf("Job %s completed successfully.", jobID)
	}

}

// Attempts to fetch the necessary informaton for a give job using an ID.
// If it fails to deliver the necessary data, may attempt to retry based on the error
// A max amount of retries is defined.
// return the full job or returns an error if it is not able to recover
func (p *JobProcessor) fetchJobWithRetry(ctx context.Context, jobID uuid.UUID) (*storage.Job, error) {

	const maxRetries = 4

	// database calls can be more frequent
	const baseDelay = 1 * time.Second

	var lastErr error

	for i := range maxRetries {

		job, err := p.db.JobByID(ctx, jobID)
		lastErr = err // Store the last error encountered

		if err == nil {
			return job, nil
		}

		if !isRetryable(err) {
			return nil, fmt.Errorf("non-retryable error fetching job %s: %w", jobID, err)
		}

		slog.Warn("Unrecoverable error fetching job, retrying...",
			"jobID", jobID, "attempt", i+1, "error", err)

		if i < maxRetries-1 {
			delay := calculateExponentialBackoff(i, baseDelay)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	// Final Failure: Max attempts reached
	return nil, fmt.Errorf("failed to fetch job %s after %d retries: %w", jobID, maxRetries, lastErr)
}

//   starts the processing of the actual job
//
// The downloaded resume has its text extracted
func (p *JobProcessor) processJobFile(ctx context.Context, job *storage.Job, jobID uuid.UUID) (string, error) {

	if err := p.updateJobWithRetry(ctx, jobID, storage.Processing); err != nil {
		return "", fmt.Errorf("Failed to update job status for job %s, with %v", jobID, err)
	}

	resume, err := p.store.Download(ctx, p.s3Bucket, job.FileUrl)

	if err != nil {
		return "", fmt.Errorf("Failed to load file with for job %s, with %v", jobID, err)
	}

	extractedText, err := p.extractTextFromPDF(resume)

	if err != nil {
		// PDF parsing errors are often permanent for a corrupted file
		return "", fmt.Errorf("An error occured while extracted text on job %s, with %v: %w", jobID, err, apperrors.ErrPermanentFailure)
	}
	return extractedText, nil
}

// attempts to update the status of the job with a given id and status
// A max number of retries is defined as well as a base delay
// function updates the status of a job or returns an error if the job is not recoverable
func (p *JobProcessor) updateJobWithRetry(ctx context.Context, jobID uuid.UUID, jobStatus storage.JobStatus) error {

	const maxRetries = 4
	const baseDelay = 1 * time.Second

	var lastErr error

	for i := range maxRetries {

		err := p.db.UpdateJobStatus(ctx, jobID, jobStatus)
		lastErr = err

		if err == nil || !isRetryable(err) {
			return err
		}

		log.Printf("Retrying failed job %s, (attempt %d). With error: %v", jobID, i+1, err)

		if i < maxRetries-1 {
			delay := calculateExponentialBackoff(i, baseDelay)
			select {
			case <-ctx.Done():
				return ctx.Err() 
			case <-time.After(delay):
				// continue to next attempt
			}
		}
	}

	return fmt.Errorf("failed to update job status for job %s after %d retries: %w", jobID, maxRetries, lastErr)
}

//   the final step to processing a job
//
// saves results of the job or returns an error if not recoverable
func (p *JobProcessor) saveResultsWithRetry(ctx context.Context, jobID uuid.UUID, content string) error {

	const maxRetries = 4

	const baseDelay = 10 * time.Second

	var lastErr error

	for i := range maxRetries {

		err := p.db.SetContentWithID(ctx, jobID, content)
		lastErr = err

		if err == nil {
			return nil // Return immediately on success.
		}

		if !isRetryable(err) {
			return fmt.Errorf("Failed to update job status for job: %s, with: %v", jobID, err)
		}

		slog.Warn("error saving resume content, retrying...",
			"jobID", jobID, "attempt", i+1, "error", err)

		if i < maxRetries-1 {
			delay := calculateExponentialBackoff(i, baseDelay)

			select {

			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):

			}
		}
	}
	return fmt.Errorf("failed to save resume content for job %s after %d retries: %w", jobID, maxRetries, lastErr)
}

// function checks whether the error is recoverable based on some common cases
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// do not retry on context cancellation
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// do not retry unrecoverable errors
	if errors.Is(err, apperrors.ErrPermanentFailure) {
		return false
	}

	return true

}

// calulates the exponential backoff by baseDelay * 2^attempts
func calculateExponentialBackoff(attempts int, baseDelay time.Duration) time.Duration {
	factor := math.Pow(2, float64(attempts))
	return time.Duration(float64(baseDelay) * factor)
}

// Scanned pdfs cause problems for general pdf reading libraries
func (p *JobProcessor) extractTextFromPDF(file []byte) (string, error) {
	const replacementChar = string(unicode.ReplacementChar)

	r := bytes.NewReader(file)
	reader, err := pdf.NewReader(r, int64(len(file)))
	if err != nil {
		return "", err
	}

	var result strings.Builder
	numPages := reader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		content := page.Content()
		texts := content.Text

		sort.Slice(texts, func(i, j int) bool {
			if texts[i].Y == texts[j].Y {
				return texts[i].X < texts[j].X
			}
			return texts[i].Y > texts[j].Y
		})

		for _, t := range texts {
			if utf8.ValidString(t.S) {
				result.WriteString(t.S)
			}
		}
		result.WriteString("\n") // Add a space or newline between pages
	}

	return strings.ReplaceAll(result.String(), replacementChar, ""), nil
}
