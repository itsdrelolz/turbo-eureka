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
	Client     *s3.Client
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
		return nil, fmt.Errorf("failed to load s3 config: %w", err)
	}

	if conf.EndpointURL != "" {
		cfg.BaseEndpoint = aws.String(conf.EndpointURL)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})


	return &FileStore{Client: client}, nil

}

func (fs *FileStore) Upload(ctx context.Context, file io.Reader, bucket, key, contentType string) (string, error) {

	_, err := fs.Client.PutObject(ctx,&s3.PutObjectInput{ 
		Bucket: aws.String(bucket), 
		Key: aws.String(key), 
		Body: file, 
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	location := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)
	return location, nil
}

func (fs *FileStore) Download(ctx context.Context, bucket, key string) ([]byte, error) { 


	result, err := fs.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket), 
		Key: aws.String(key),
	})

	if err != nil { 
		return nil, fmt.Errorf("failed to download file: %w", err) 
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read object body from file: %w", err)
	}

	return body, nil

}
