package mocks

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"job-matcher/internal/storage"
)

type MockJobCreator struct {
	mock.Mock
}

func (m *MockJobCreator) InsertJobReturnID(ctx context.Context, fileUrl string, jobStatus storage.JobStatus) (uuid.UUID, error) {
	args := m.Called(ctx, fileUrl, jobStatus)

	return args.Get(0).(uuid.UUID), args.Error(1)
}
