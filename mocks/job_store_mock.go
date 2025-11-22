package mocks

import (
	"context"
	"job-matcher/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockCreator struct {
	mock.Mock
}

func (m *MockCreator) Create(ctx context.Context, fileName string, jobStatus models.JobStatus) (uuid.UUID, error) {
	args := m.Called(ctx, fileName, jobStatus)

	return args.Get(0).(uuid.UUID), args.Error(1)
}
