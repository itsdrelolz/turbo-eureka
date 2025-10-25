package objectstore

import ( 
	"io"
	"context" 
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
)


type FileStorer interface { 
	Upload(ctx context.Context, file io.Reader, bucket, key, contentType string) (*manager.UploadOutput, error)
}
