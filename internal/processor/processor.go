package processor

import (
	"context"
	"io"
	"job-matcher/internal/models"
	"log"
	"sync"
	"time"
	"rsc.io/pdf"
	"github.com/google/uuid"
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

	jobChan := make(chan uuid.UUID, buffLen)

	var wg sync.WaitGroup

	go p.startConsumer(ctx, jobChan)

	for i := 1; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			p.startWorkers(ctx, workerID, jobChan)
		}(i)
	}

	<-ctx.Done()
	log.Printf("Shutdown signal received. Waiting for workers to finish...")

	wg.Wait()
	log.Printf("All workers stopped. Job processor shut down")
}

func (p *JobProcessor) startConsumer(ctx context.Context, jobChan chan<- uuid.UUID) {

	defer close(jobChan)

	log.Printf("Job processor has started, waiting for jobs...")

	for {
		jobID, err := p.queue.Consume(ctx)

		if err != nil {
			log.Printf("error consuming from job queue with err: %w, retrying..", err)
			time.Sleep(5 * time.Second)
			continue
		}
		select {
		// job successfully sent to worker
		case jobChan <- jobID:
		case <-ctx.Done():
			log.Printf("Job consumer stopping")
			return
		}
	}
}

func (p *JobProcessor) startWorkers(ctx context.Context, workerID int, jobChan <-chan uuid.UUID) {

	log.Printf("Worker: #%d started", workerID)

	for jobID := range jobChan {
		p.processJob(ctx, jobID)
	}
	log.Printf("Worker #%d stopped.", workerID)

}

func (p *JobProcessor) processJob(ctx context.Context, jobID uuid.UUID) {



	err := p.db.ProcessingJob(ctx, jobID) 

	if err != nil { 
		log.Printf("ERROR: %w", err)
	}

	jobInfo, err := p.db.Get(ctx, jobID) 

	if err != nil { 
		log.Printf("ERROR: %w", err) 
	}


	resume, err := p.s3.Download(ctx, p.bucket, jobInfo.FileName)

	if err != nil { 
		log.Printf("ERROR: %w", err)
	}


}


func (p *JobProcessor) extractText(ctx context.Context, resume io.ReadCloser) {



}
