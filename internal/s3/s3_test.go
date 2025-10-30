package s3_test

import (
	"bytes"
	"context"
	"job-matcher/internal/s3"
	"os"
	"testing"

	"github.com/google/uuid"
)

func setUpS3(t *testing.T) (*s3.FileStore, string) {
	t.Helper()

	// Get configuration from environment variables
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")

	if endpoint == "" || accessKey == "" || secretKey == "" {
		t.Skip("MinIO configuration not set (MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY), skipping integration test")
	}

	if bucket == "" {
		bucket = "resume-bucket"
	}

	ctx := context.Background()

	s3Config := s3.S3Config{
		EndpointURL: endpoint,
		Region:      "us-east-1",
		AccessKey:   accessKey,
		SecretKey:   secretKey,
	}

	s3Store, err := s3.NewFileStore(ctx, s3Config)
	if err != nil {
		t.Fatalf("Failed creating FileStore: %v", err)
	}

	return s3Store, bucket
}

// TestUploadSuccess verifies PDF file upload to MinIO
func TestUploadSuccess(t *testing.T) {
	s3Store, bucket := setUpS3(t)
	ctx := context.Background()

	// Create mock PDF content with valid PDF header
	mockPDFContent := []byte("%PDF-1.4\n%Mock PDF content for testing\n%%EOF")
	fileReader := bytes.NewReader(mockPDFContent)

	// Generate unique key to avoid collisions
	key := "test-resumes/test-resume-" + uuid.New().String() + ".pdf"
	contentType := "application/pdf"

	// Upload file
	output, err := s3Store.Upload(ctx, fileReader, bucket, key, contentType)
	if err != nil {
		t.Fatalf("Failed to upload file: %v", err)
	}

	// Verify output
	if output == "" {
		t.Fatal("Expected non-nil upload output, got nil")
	}

	t.Logf("Successfully uploaded file to: %s", output)
}

// TestUploadMultiplePDFs tests uploading multiple PDF files
func TestUploadMultiplePDFs(t *testing.T) {
	s3Store, bucket := setUpS3(t)
	ctx := context.Background()

	testCases := []struct {
		name        string
		content     string
		contentType string
	}{
		{
			name:        "small-pdf",
			content:     "%PDF-1.4\nSmall test PDF\n%%EOF",
			contentType: "application/pdf",
		},
		{
			name:        "large-pdf",
			content:     "%PDF-1.4\n" + string(make([]byte, 1024*100)) + "\n%%EOF",
			contentType: "application/pdf",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileReader := bytes.NewReader([]byte(tc.content))
			key := "test-resumes/" + tc.name + "-" + uuid.New().String() + ".pdf"

			output, err := s3Store.Upload(ctx, fileReader, bucket, key, tc.contentType)
			if err != nil {
				t.Errorf("Failed to upload %s: %v", tc.name, err)
			}

			if output == "" {
				t.Errorf("Invalid upload output for %s", tc.name)
			}
		})
	}
}

// TestUploadInvalidBucket tests upload to non-existent bucket
func TestUploadInvalidBucket(t *testing.T) {
	s3Store, _ := setUpS3(t)
	ctx := context.Background()

	mockPDFContent := []byte("%PDF-1.4\nTest content\n%%EOF")
	fileReader := bytes.NewReader(mockPDFContent)

	invalidBucket := "non-existent-bucket-" + uuid.New().String()
	key := "test-file.pdf"

	_, err := s3Store.Upload(ctx, fileReader, invalidBucket, key, "application/pdf")
	if err == nil {
		t.Error("Expected error when uploading to non-existent bucket, got nil")
	}
}

// TestUploadEmptyFile tests uploading an empty PDF
func TestUploadEmptyFile(t *testing.T) {
	s3Store, bucket := setUpS3(t)
	ctx := context.Background()

	// Empty reader
	fileReader := bytes.NewReader([]byte{})
	key := "test-resumes/empty-" + uuid.New().String() + ".pdf"

	output, err := s3Store.Upload(ctx, fileReader, bucket, key, "application/pdf")
	if err != nil {
		t.Logf("Upload empty file returned error (expected): %v", err)
	}

	// Some implementations may allow empty files, others may not
	// Just verify we get a response
	if err == nil && output == "" {
		t.Error("Expected either error or output, got neither")
	}
}
