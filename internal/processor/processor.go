package processor

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"io"
	"job-matcher/internal/models"
	"log"
	"rsc.io/pdf"
	"strings"
	"sync"
	"unicode"
)

const (
	numWorkers = 5
)

type JobMetaData struct { 
	ID uuid.UUID `json:"id"`
	S3Key string `json:"s3_key"`
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

func (p *JobProcessor) Run(ctx context.Context) { 


	result := make(chan JobMetaData, numWorkers * 2) 
	
	var wg sync.WaitGroup

	

	for i := 1; i <= numWorkers; i++ {

	wg.Go(func() { 
		p.startWorker(ctx, result)
	})
	}

	wg.Wait()


	wg.Done()

}



func (p *JobProcessor) startWorker(ctx context.Context, result <-chan JobMetaData) { 
	

}

func (p *JobProcessor) startConsumer(ctx context.Context, result <- 




func (p *JobProcessor) extractText(resume io.ReadCloser) (string, error) {

	// identifier for unreadable character common in output
	// this character will be removed from the output
	const replacementChar = string(unicode.ReplacementChar)

	fileBytes, err := io.ReadAll(resume)

	if err != nil {
		return "", fmt.Errorf("Failed to read file stream: %w", err)
	}


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
