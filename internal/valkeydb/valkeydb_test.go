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
        t.Fatalf("Failed to insert job: %v", err)
    }


	// USE Blocking Right pop to simulate workers pulling from queue
    brpopCmd := valkeyDB.Client.B().Brpop().
        Key("job-queue").
        Timeout(0).
        Build()

    res := valkeyDB.Client.Do(ctx, brpopCmd)

    arr, err := res.AsStrSlice()
    if err != nil {
        t.Fatalf("Failed to parse BRPOP response: %v", err)
    }
    if len(arr) != 2 {
        t.Fatalf("Expected 2 elements in BRPOP response, got %d", len(arr))
    }

    poppedJobID := arr[1]

    if poppedJobID != jobID {
        t.Fatalf("Expected job ID %s, got %s", jobID, poppedJobID)
    }
}

