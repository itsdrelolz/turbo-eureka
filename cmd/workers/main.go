package main

import (
	"context"
	"github.com/joho/godotenv"
	"job-matcher/internal/postgres"
	"job-matcher/internal/processor"
	"job-matcher/internal/redis"
	"job-matcher/internal/s3"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	ctx := context.Background()
	postgresDB, err := postgresdb.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatalln("Failed to initialize postgresdb")
	}
	defer postgresDB.Close()

	redis, err := redis.New(ctx, os.Getenv("REDIS_URL"), os.Getenv("REDIS_PASSWORD"))

	if err != nil {
		log.Fatalln("Failed to initialize redis")
	}

	defer redis.Client.Close()

	s3Conf := s3.S3Config{
		EndpointURL: os.Getenv("S3_ENDPOINT_URL"),
		Region:      os.Getenv("S3_REGION"),
		AccessKey:   os.Getenv("S3_ACCESS_KEY"),
		SecretKey:   os.Getenv("S3_SECRET_KEY"),
	}

	s3Store, err := s3.NewFileStore(ctx, s3Conf)

	if err != nil {
		log.Fatalln("Could not create S3 filestore")
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("S3_BUCKET_NAME is not set")
	}

	workerQueue := processor.NewJobProcessor(
		postgresDB,
		redis,
		s3Store,
		bucketName,
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
