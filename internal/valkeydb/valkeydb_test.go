package valkeydb_test

import (
	"context"
	"job-matcher/internal/valkeydb"
	"os"
	"testing"
)

func setUpTestDB(t *testing.T) *valkeydb.ValkeyClient {

	t.Helper()

	url := os.Getenv("VALKEY_TEST_URL")
	if url == "" {
		t.Skip("VALKEY_TEST_URL not set, skipping integration test")
	}

	valkeypass := os.Getenv("VALKEY_TEST_PASSWORD")
	if valkeypass == "" {
		t.Skip("VALKEY_PASSWORD not set, skipping integration test")
	}
	ctx := context.Background()

	db, err := valkeydb.New(ctx, url, valkeypass)

	if err != nil {

		t.Fatalf("failed to connect to test database: %v", err)
	}

	return db
}

func TestInsertJobSuccess(t *testing.T) {
	valkeyDB := setUpTestDB(t)
	ctx := context.Background()
	jobID := "test_job_id"

	err := valkeyDB.InsertJob(ctx, jobID)
	if err != nil {
		t.Fatalf("Failed to insert job with err: %v", err)
	}

	popCmd := valkeyDB.Client.B().Lpop().Key("job-queue").Build()
	res := valkeyDB.Client.Do(ctx, popCmd)

	poppedJobID, err := res.ToString()
	if err != nil {
		t.Fatalf("Failed to retrieve job from queue: %v", err)
	}

	if poppedJobID != jobID {
		t.Fatalf("Expected job ID %s, got %s", jobID, poppedJobID)
	}
}
