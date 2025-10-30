package objectstore

import (
	"context"
	"io"

)

type FileStorer interface {
	Upload(ctx context.Context, file io.Reader, bucket, key, contentType string) (string, error)
	Download(ctx context.Context, bucket, key string) ([]byte, error)
}
