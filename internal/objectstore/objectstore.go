package objectstore

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"io"
)

type FileStorer interface {
	Upload(ctx context.Context, file io.Reader, bucket, key, contentType string) (*manager.UploadOutput, error)
}
