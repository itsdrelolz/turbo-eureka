package main

import (
	"context"
	"job-matcher/internal/geministore"
	"job-matcher/internal/postgresdb"
	"job-matcher/internal/processor"
	"job-matcher/internal/s3"
	"job-matcher/internal/valkeydb"
	"log"
	"os"
)

func main() {

	ctx := context.Background()
	postgres, err := postgresdb.New(ctx, os.Getenv("DB_URL"))

	if err != nil {
		log.Fatalf("failed database initialization with err: %w", err)
	}

	valkey, err := valkeydb.New(ctx, os.Getenv("VALKEY_URL"), os.Getenv("VALKEY_PASSWORD"))

	if err != nil {
		log.Fatalf("failed valkey initialization with err: %w", err)
	}



	s3, err := s3.NewFileStore()


	if err != nil { 
		log.Fatalf("failed s3 initialization with err: %w", err)
	}
	gemini, err := geministore.New(ctx, os.Getenv("GEMINI_API_KEY"))

	if err != nil { 
		log.Fatalf("no gemini API key given")
	}

	workerQueue := processor.NewJobProcessor(postgres, valkey, s3, gemini)


}
