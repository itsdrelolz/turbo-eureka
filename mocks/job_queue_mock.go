package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockJobQueuer struct {
	mock.Mock
}

func (m *MockJobQueuer) InsertJob(ctx context.Context, jobID string) error {
	args := m.Called(ctx, jobID)

	return args.Error(0)
}
