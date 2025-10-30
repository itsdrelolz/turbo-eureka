package mocks

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

type MockFileStorer struct {
	mock.Mock
}

func (m *MockFileStorer) Upload(ctx context.Context, file io.Reader, bucket, key, contentType string) (string, error) {
	args := m.Called(ctx, file, bucket, key, contentType)

	if args.Get(0) == nil {
		return "", args.Error(1)
	}

	return args.Get(0).(string), args.Error(1)
}

func (m *MockFileStorer) Download(ctx context.Context, bucket, key string) ([]byte, error) {

	args := m.Called(ctx, bucket, key)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)

}
