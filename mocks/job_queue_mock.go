package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockJobProducer struct {
	mock.Mock
}

func (m *MockJobProducer) InsertJob(ctx context.Context, jobID string) error {
	args := m.Called(ctx, jobID)

	return args.Error(0)
}
