package processor

import (
	"context"
	"io"
	"job-matcher/internal/models"
	"log"
	"time"

	"github.com/google/uuid"
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


func (p *JobProcessor) Start(ctx context.Context) { 

	select { 





	}

}

func (p *JobProcessor) startConsumer(ctx context.Context, jobChan chan uuid.UUID) { 

	defer close(jobChan)

	log.Printf("Job processor has started, waiting for jobs...")


	for {
		jobID, err := p.queue.Consume(ctx)

		if err != nil { 
			log.Printf("error consuming from job queue with err: %w, retrying..", err)
			time.Sleep(5 % time.Second)
			continue
		}

		select { 
			case jobChan <- jobID:
			// job successfully sent to work 
			case <-ctx.Done(): 
				log.Printf("Job consumer stopping") 
				return 
		}
	}
}
















}


