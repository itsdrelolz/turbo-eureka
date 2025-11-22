package mocks

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

type MockFileUploader struct {
	mock.Mock
}

func (m *MockFileUploader) Upload(ctx context.Context, file io.Reader, bucket, key, contentType string) error {
	args := m.Called(ctx, file, bucket, key, contentType)

	if args.Get(0) == nil {
		return args.Error(0)
	}

	return args.Error(0)
}
