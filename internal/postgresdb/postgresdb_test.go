package postgresdb_test

import (
	"context"
	"job-matcher/internal/models"
	"job-matcher/internal/postgresdb"
	"os"
	"testing"

	"github.com/google/uuid"
)

func setUpTestDB(t *testing.T) *postgresdb.Store {

	t.Helper()

	connString := os.Getenv("DATABASE_URL")

	if connString == "" {
		t.Skip("DB_TEST_URL not set, skipping integration test")
	}

	ctx := context.Background()

	db, err := postgresdb.New(ctx, connString)

	if err != nil {

		t.Fatalf("failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {

		_, err := db.Pool.Exec(ctx, "TRUNCATE TABLE jobs")
		if err != nil {

			t.Fatalf(" Failed to clean up jobs table: %v", err)
		}

		db.Close()
	})
	return db
}

func TestCreateSuccess(t *testing.T) {

	postgresDB := setUpTestDB(t)

	ctx := context.Background()

	resumeName := "https://example.com/resumes/johndoe.pdf"

	jobID, err := postgresDB.Create(ctx, resumeName, models.Queued)

	if err != nil {

		t.Fatalf("InsertJobID() returned an unexpected error: %v", err)
	}

	if err := uuid.Validate(jobID.String()); err != nil {
		t.Fatalf("InsertJobID() returned an invalid ID: got %v, want a non-empty string", jobID)
	}

	var savedURL string

	query := "SELECT file_name FROM jobs where id=$1"

	err = postgresDB.Pool.QueryRow(ctx, query, jobID).Scan(&savedURL)

	if err != nil {

		t.Fatalf("Failed to query back the inserted job: %v", err)
	}

	if savedURL != resumeName {

		t.Errorf("URL in database does not match: got %q, want %q", savedURL, resumeName)
	}
}
