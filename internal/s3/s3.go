package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

type FileStore struct {
	Client *s3.Client
}

type S3Config struct {
	EndpointURL string
	Region      string
	AccessKey   string
	SecretKey   string
}

func NewFileStore(ctx context.Context, conf S3Config) (*FileStore, error) {

	creds := credentials.NewStaticCredentialsProvider(conf.AccessKey, conf.SecretKey, "")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(conf.Region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to load s3 config: %w", err)
	}

	if conf.EndpointURL != "" {
		cfg.BaseEndpoint = aws.String(conf.EndpointURL)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &FileStore{Client: client}, nil

}

func (fs *FileStore) Upload(ctx context.Context, file io.Reader, bucket, key, contentType string) error {

	if file == nil {
		return fmt.Errorf("ERROR: empty file, invalid input")
	}

	_, err := fs.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		ChecksumAlgorithm: "SHA256",
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return fmt.Errorf("ERROR: failed to upload file: %w", err)
	}
	return nil
}

func (fs *FileStore) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {

	result, err := fs.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to download file: %w", err)
	}
	defer result.Body.Close()

	return result.Body, nil
}

// TODO this function is meant to use the checksum of each object in the s3 bucket and compare with the users newly uploaded file 
// If the checksum matches, return true if the file is already uploaded to the bucket, preventing unecessary duplicate extractions 

func (fs *FileStore) IsDuplicate(ctx context.Context, bucket, key string) bool { 



}
