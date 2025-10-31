package mocks

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"job-matcher/internal/storage"
)

type MockJobStore struct {
	mock.Mock
}

func (m *MockJobStore) InsertJobReturnID(ctx context.Context, fileUrl string, jobStatus storage.JobStatus) (uuid.UUID, error) {
	args := m.Called(ctx, fileUrl, jobStatus)

	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockJobStore) JobByID(ctx context.Context, jobID uuid.UUID) (*storage.Job, error) {
	args := m.Called(ctx, jobID)

	if args.Get(0) == nil {
		return &storage.Job{}, args.Error(1)
	}

	return args.Get(0).(*storage.Job), args.Error(1)
}

func (m *MockJobStore) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, jobStatus storage.JobStatus) error {
	args := m.Called(ctx, jobID, jobStatus)
	return args.Error(1)
}
