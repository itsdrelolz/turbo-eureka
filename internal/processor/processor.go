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
	"unicode"
	"github.com/google/uuid"
	"rsc.io/pdf"
)

const (
	numWorkers = 5
)

type EventType int

const (
	JobFailed EventType = iota
	JobCompleted
)

type JobEvent struct {
	JobID uuid.UUID
	Type  EventType
	Data  string
	Err   error
}

type Consumer interface {
	Consume(ctx context.Context) (uuid.UUID, error)
}

type Downloader interface {
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)
}

type JobStore interface {
	Get(ctx context.Context, jobID uuid.UUID) (*models.Job, error)
	CompleteJob(ctx context.Context, jobID uuid.UUID, data string) error
	FailJob(ctx context.Context, jobID uuid.UUID, errMsg error) error 
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

// Function implements two waitgroups, one for the main workers, and one for handling db writes
// this prevents potential dataloss by allowing the database transactions to finish before closing the channel
func (p *JobProcessor) Run(ctx context.Context) {

	results := make(chan JobEvent, numWorkers)

	var workerWg sync.WaitGroup
	var aggWg sync.WaitGroup

	aggWg.Go(func() {
		p.aggregator(results)
	})

	for i := 0; i < numWorkers; i++ {
		workerWg.Go(func() {
			p.startWorker(ctx, results)
		})
	}
	workerWg.Wait()

	close(results)

	aggWg.Wait()
}

func (p *JobProcessor) startWorker(ctx context.Context, result chan<- JobEvent) {

	for {
		select {
		case <-ctx.Done():
			return
		default:
			jobID, err := p.queue.Consume(ctx)

			if err != nil {
				log.Printf("Failed to pull job from queue with err: %w", err)
			}

			jobData, err := p.db.Get(ctx, jobID)

			if err != nil {
				result <- JobEvent{JobID: jobID, Type: JobFailed, Err: err}
				continue
			}

			resume, err := p.s3.Download(ctx, p.bucket, jobData.FileName)

			if err != nil {
				result <- JobEvent{JobID: jobID, Type: JobFailed, Err: err}
				continue
			}

			text, err := p.extractText(resume)

			if err != nil {
				result <- JobEvent{JobID: jobID, Type: JobFailed, Err: err}
				continue
			}

			result <- JobEvent{JobID: jobID, Type: JobCompleted, Data: text}
		}
	}
}

func (p *JobProcessor) aggregator(events <-chan JobEvent) {

	ctx := context.Background()

	for evt := range events {
		switch evt.Type {
		case JobCompleted:
			p.db.CompleteJob(ctx, evt.JobID, evt.Data)

		case JobFailed:
			p.db.FailJob(ctx, evt.JobID, evt.Err)
		}
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
