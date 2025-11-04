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
	"os/signal"
	"syscall"
)

func main() {

	ctx := context.Background()
	postgresDB, err := postgresdb.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatalf("Failed to initialize postgresdb: %v", err)
	}
	defer postgresDB.Close()

	valkeyQueue, err := valkeydb.New(ctx, os.Getenv("VALKEY_URL"), os.Getenv("VALKEY_PASSWORD"))

	if err != nil {
		log.Fatalf("Failed to initialize valkey: %v", err)
	}

	defer valkeyQueue.Close()

	s3Conf := s3.S3Config{
		EndpointURL: os.Getenv("S3_ENDPOINT_URL"),
		Region:      os.Getenv("S3_REGION"),
		AccessKey:   os.Getenv("S3_ACCESS_KEY"),
		SecretKey:   os.Getenv("S3_SECRET_KEY"),
	}

	s3Store, err := s3.NewFileStore(ctx, s3Conf)

	if err != nil {
		log.Fatalf("Could not create S3 filestore: %v", err)
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("S3_BUCKET_NAME is not set")
	}

	gemini, err := geministore.New(ctx, os.Getenv("GEMINI_API_KEY"))

	if err != nil { 
		log.Fatalf("no gemini API key given")
	}

	workerQueue := processor.NewJobProcessor(
		postgresDB, 
		valkeyQueue, 
		s3Store, 
		bucketName, 
		gemini,
		)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go workerQueue.Run(ctx)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping workers...")
	cancel() 

	log.Println("Worker shutdown complete")
}

