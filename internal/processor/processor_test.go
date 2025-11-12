package processor

import (
	"context"
	"errors"
	"github.com/google/uuid"
	apperrors "job-matcher/internal/errors"
	"job-matcher/internal/storage"
	"strings"
	"testing"
	"time"
)

type MockJobStore struct {
	JobByIDFunc          func(ctx context.Context, id uuid.UUID) (*storage.Job, error)
	UpdateJobStatusFunc  func(ctx context.Context, id uuid.UUID, status storage.JobStatus) error
	SetContentWithIDFunc func(ctx context.Context, id uuid.UUID, content string) error
}

func (m *MockJobStore) JobByID(ctx context.Context, jobID uuid.UUID) (*storage.Job, error) {
	return m.JobByIDFunc(ctx, jobID)
}

func (m *MockJobStore) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, jobStatus storage.JobStatus) error {
	return m.UpdateJobStatusFunc(ctx, jobID, jobStatus)
}

func (m *MockJobStore) SetContentWithID(ctx context.Context, jobID uuid.UUID, resumeContent string) error {
	return m.SetContentWithIDFunc(ctx, jobID, resumeContent)
}

type MockFileFetcher struct {
	DownloadFunc func(ctx context.Context, bucket, key string) ([]byte, error)
}

func (m *MockFileFetcher) Download(ctx context.Context, bucket, key string) ([]byte, error) {
	return m.DownloadFunc(ctx, bucket, key)
}

func TestIsRetryable(t *testing.T) {
	if isRetryable(context.DeadlineExceeded) {
		t.Errorf("Deadline exceeded should not be retryable")
	}
	if isRetryable(context.Canceled) {
		t.Errorf("Canceled context should not be retryable")
	}
	if isRetryable(apperrors.ErrPermanentFailure) {
		t.Errorf("ErrPermanentFailure should not be retryable")
	}
	if !isRetryable(errors.New("connection failed")) {
		t.Errorf("Generic error should be retryable")
	}
}

func TestExponentialBackoff(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	tests := []struct {
		attempts int
		expected time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
	}
	for _, tt := range tests {
		delay := calculateExponentialBackoff(tt.attempts, baseDelay)
		if delay != tt.expected {
			t.Errorf("Attempt %d: Got %v, want %v", tt.attempts, delay, tt.expected)
		}
	}
}

func TestFetchJobWithRetrySuccessOnFirstAttempt(t *testing.T) {
	jobID := uuid.New()
	expectedJob := &storage.Job{ID: jobID, FileName: "test/file.pdf"}
	calls := 0
	mockDB := &MockJobStore{
		JobByIDFunc: func(ctx context.Context, id uuid.UUID) (*storage.Job, error) {
			calls++
			return expectedJob, nil
		},
	}
	processor := &JobProcessor{db: mockDB}
	job, err := processor.fetchJobWithRetry(context.Background(), jobID)
	if err != nil {
		t.Fatalf("fetchJobWithRetry failed unexpectedly: %v", err)
	}
	if job == nil || job.ID != jobID {
		t.Errorf("Did not return the expected job. Got: %v", job)
	}
	if calls != 1 {
		t.Errorf("Expected 1 DB call, got %d", calls)
	}
}

func TestFetchJobWithRetryFailsAfterMaxRetries(t *testing.T) {
	jobID := uuid.New()
	attempts := 0
	retryableErr := errors.New("temporary DB connection issue")
	mockDB := &MockJobStore{
		JobByIDFunc: func(ctx context.Context, id uuid.UUID) (*storage.Job, error) {
			attempts++
			return nil, retryableErr
		},
	}
	processor := &JobProcessor{db: mockDB}
	job, err := processor.fetchJobWithRetry(context.Background(), jobID)
	if attempts != 4 {
		t.Errorf("Expected 4 attempts, got %d", attempts)
	}
	if err == nil {
		t.Fatal("fetchJobWithRetry expected to fail, but succeeded")
	}
	if job != nil {
		t.Errorf("Expected nil job on failure, got %v", job)
	}
	if !strings.Contains(err.Error(), "failed to fetch job") {
		t.Errorf("Did not receive expected final failure error: %v", err)
	}
	if !errors.Is(err, retryableErr) {
		t.Errorf("Final error should wrap the last temporary error: got %v", err)
	}
}

func TestUpdateJobWithRetryRetriesAndSucceeds(t *testing.T) {
	jobID := uuid.New()
	calls := 0
	mockDB := &MockJobStore{
		UpdateJobStatusFunc: func(ctx context.Context, id uuid.UUID, status storage.JobStatus) error {
			calls++
			if calls < 3 {
				return errors.New("temp connection loss")
			}
			return nil
		},
	}
	processor := &JobProcessor{db: mockDB}
	err := processor.updateJobWithRetry(context.Background(), jobID, storage.Completed)
	if err != nil {
		t.Fatalf("updateJobWithRetry failed unexpectedly: %v", err)
	}
	if calls != 3 {
		t.Errorf("Expected 3 DB calls, got %d", calls)
	}
}

func TestUpdateJobWithRetryPermanentFailure(t *testing.T) {
	jobID := uuid.New()
	calls := 0
	mockDB := &MockJobStore{
		UpdateJobStatusFunc: func(ctx context.Context, id uuid.UUID, status storage.JobStatus) error {
			calls++
			return apperrors.ErrPermanentFailure
		},
	}
	processor := &JobProcessor{db: mockDB}
	err := processor.updateJobWithRetry(context.Background(), jobID, storage.Processing)
	if err == nil {
		t.Fatal("updateJobWithRetry expected to fail permanently, but succeeded")
	}
	if calls != 1 {
		t.Errorf("Expected 1 DB call on permanent failure, got %d", calls)
	}
	if !errors.Is(err, apperrors.ErrPermanentFailure) {
		t.Errorf("Error was not the expected permanent failure: %v", err)
	}
}
