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

func (m *MockJobStore) InsertJobAndGetID(ctx context.Context, fileUrl string) (uuid.NullUUID, error) {
	args := m.Called(ctx, fileUrl)

	return args.Get(0).(uuid.NullUUID), args.Error(1)
}

func (m *MockJobStore) GetJobByID(ctx context.Context, jobID uuid.NullUUID) (storage.Job, error) {
	args := m.Called(ctx, jobID)

	if args.Get(0) == nil {
		return storage.Job{}, args.Error(1)
	}

	return args.Get(0).(storage.Job), args.Error(1)
}
