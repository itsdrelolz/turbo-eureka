package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Produce(ctx context.Context, jobID string) error {
	args := m.Called(ctx, jobID)

	return args.Error(0)
}
