package workers

import (
	"context"
	"job-matcher/internal/postgresdb"
	"job-matcher/internal/processor"
	"job-matcher/internal/s3"
	"job-matcher/internal/valkeydb"
	"log"
	"os"
)

func main() {

	ctx := context.Background()
	db, err := postgresdb.New(ctx, os.Getenv("DB_URL"))

	if err != nil {
		log.Fatalf("failed database initialization with err: %w", err)
	}

	valkey, err := valkeydb.New(ctx, os.Getenv("VALKEY_URL"), os.Getenv("VALKEY_PASSWORD"))

	if err != nil {
		log.Fatalf("failed valkey initialization with err: %w", err)
	}

}
