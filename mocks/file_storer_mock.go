package mocks

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/stretchr/testify/mock"
)

type MockFileStorer struct {
	mock.Mock
}

func (m *MockFileStorer) Upload(ctx context.Context, file io.Reader, bucket, key, contentType string) (*manager.UploadOutput, error) {
	args := m.Called(ctx, file, bucket, key, contentType)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*manager.UploadOutput), args.Error(1)
}
