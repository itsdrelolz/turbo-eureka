package processor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"job-matcher/internal/models"
	"log"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
	"rsc.io/pdf"
)

const (
	numWorkers = 5
	buffLen    = numWorkers * 2
)

type Consumer interface {
	Consume(ctx context.Context) (uuid.UUID, error)
}

type Downloader interface {
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)
}

type JobStore interface {
	ProcessingJob(ctx context.Context, jobID uuid.UUID) error

	FailJob(ctx context.Context, jobID uuid.UUID) error

	CompleteJob(ctx context.Context, jobID uuid.UUID, content string) error

	Get(ctx context.Context, jobID uuid.UUID) (*models.Job, error)
}
type JobProcessor struct {
	db     JobStore
	queue  Consumer
	s3     Downloader
	bucket string
}

func NewJobProcessor(db JobStore, queue Consumer, s3 Downloader, bucket string) *JobProcessor {
	return &JobProcessor{db: db, queue: queue, s3: s3, bucket: bucket}
}

// TODO: Runner function for the job processor functionality
// This function should start other functions

func (p *JobProcessor) Run(ctx context.Context) {

	workerQueue := make(chan uuid.UUID, buffLen)

	var wg sync.WaitGroup

	go p.startConsumer(ctx, workerQueue)

	for i := 1; i < numWorkers; i++ {
		wg.Go(func() {
			p.startWorkers(ctx, i, workerQueue)
		})
	}

	<-ctx.Done()
	log.Printf("Shutdown signal received. Waiting for workers to finish...")

	wg.Wait()
	log.Printf("All workers stopped. Job processor shut down")
}

func (p *JobProcessor) startConsumer(ctx context.Context, workerQueue chan<- uuid.UUID) {

	defer close(workerQueue)

	log.Printf("Job processor has started, waiting for jobs...")

	for {
		jobID, err := p.queue.Consume(ctx)

		if err != nil {
			log.Printf("error consuming from job queue with err: %v, retrying..", err)
			time.Sleep(5 * time.Second)
			continue
		}
		select {
		// job successfully sent to worker
		case workerQueue <- jobID:
		case <-ctx.Done():
			log.Printf("Job consumer stopping")
			return
		}
	}
}

func (p *JobProcessor) startWorkers(ctx context.Context, workerID int, workerQueue <-chan uuid.UUID) {

	log.Printf("Worker: #%d started", workerID)

	for jobID := range workerQueue {
		p.processJob(ctx, jobID)
	}
	log.Printf("Worker #%d stopped.", workerID)

}

func (p *JobProcessor) processJob(ctx context.Context, jobID uuid.UUID) {

	err := p.db.ProcessingJob(ctx, jobID)

	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	jobInfo, err := p.db.Get(ctx, jobID)

	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	resume, err := p.s3.Download(ctx, p.bucket, jobInfo.FileName)

	if err != nil {
		log.Printf("ERROR: %v", err)
		p.db.FailJob(ctx, jobID)
		return
	}

	text, err := p.extractText(resume)

	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}

	err = p.db.CompleteJob(ctx, jobID, text)

	if err != nil {
		log.Printf("ERROR: %v", err)
	}

}

func (p *JobProcessor) extractText(resume io.ReadCloser) (string, error) {

	// identifier for unreadable character common in output
	// this character will be removed from the output
	const replacementChar = string(unicode.ReplacementChar)

	fileBytes, err := io.ReadAll(resume)

	if err != nil {
		return "", fmt.Errorf("Failed to read file stream: %w", err)
	}
	defer resume.Close()

	readerAt := bytes.NewReader(fileBytes)

	reader, err := pdf.NewReader(readerAt, int64(len(fileBytes)))

	if err != nil {
		return "", fmt.Errorf("Failed to open file: %w", err)
	}

	var result strings.Builder

	numPages := reader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)

		content := page.Content()

		text := content.Text

		for _, t := range text {
			result.WriteString(t.S)
		}
		result.WriteString("\n")
	}

	// Remove all occurences of the replacment char in the resulting string
	return strings.ReplaceAll(result.String(), replacementChar, ""), nil

}

// Cleanup function
/*

func (p *JobProcessor) cleanUp(ctx context.Context, deadLetterQueue chan<- uuid.UUID) error {



}
*/
