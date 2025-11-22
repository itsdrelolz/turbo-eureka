package processor

import (
	"context"
	"errors"
	apperrors "job-matcher/internal/errors"
	"job-matcher/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type MockJobStore struct {
	GetFunc          func(ctx context.Context, id uuid.UUID) (*models.Job, error)
	UpdateStatusFunc func(ctx context.Context, id uuid.UUID, status models.JobStatus) error
	UpdateResumeFunc func(ctx context.Context, id uuid.UUID, content string) error
}

func (m *MockJobStore) Get(ctx context.Context, jobID uuid.UUID) (*models.Job, error) {
	return m.GetFunc(ctx, jobID)
}

func (m *MockJobStore) UpdateStatus(ctx context.Context, jobID uuid.UUID, jobStatus models.JobStatus) error {
	return m.UpdateStatusFunc(ctx, jobID, jobStatus)
}

func (m *MockJobStore) UpdateResume(ctx context.Context, jobID uuid.UUID, resumeContent string) error {
	return m.UpdateResumeFunc(ctx, jobID, resumeContent)
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
	expectedJob := &models.Job{ID: jobID, FileName: "test/file.pdf"}
	calls := 0
	mockDB := &MockJobStore{
		GetFunc: func(ctx context.Context, id uuid.UUID) (*models.Job, error) {
			calls++
			return expectedJob, nil
		},
	}
	processor := &JobProcessor{job: mockDB}
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
		GetFunc: func(ctx context.Context, id uuid.UUID) (*models.Job, error) {
			attempts++
			return nil, retryableErr
		},
	}
	processor := &JobProcessor{job: mockDB}
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
		UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.JobStatus) error {
			calls++
			if calls < 3 {
				return errors.New("temp connection loss")
			}
			return nil
		},
	}
	processor := &JobProcessor{job: mockDB}
	err := processor.updateJobWithRetry(context.Background(), jobID, models.Completed)
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
		UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.JobStatus) error {
			calls++
			return apperrors.ErrPermanentFailure
		},
	}
	processor := &JobProcessor{job: mockDB}
	err := processor.updateJobWithRetry(context.Background(), jobID, models.Processing)
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
