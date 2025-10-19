package main

import (
	"context"
	"job-matcher/internal/database"
	"job-matcher/internal/objectstore"
	"job-matcher/internal/queue"
	"log"
	"net/http"
	"os"
	"job-matcher/internal/api"
)

func main() {

	ctx := context.Background()
	dbStore, err := database.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		panic(err)
	}

	defer dbStore.Close()

	redisClient, err := queue.New(ctx, os.Getenv("REDIS_URL"))

	if err != nil {
		panic(err)
	}

	s3Conf := objectstore.S3Config{
		EndpointURL: os.Getenv("S3_ENDPOINT_URL"),
		Region:      os.Getenv("S3_REGION"),
		AccessKey:   os.Getenv("S3_ACCESS_KEY"),
		SecretKey:   os.Getenv("S3_SECRET_KEY"),
	}

	// 2. Create the FileStore
	fileStore, err := objectstore.NewFileStore(ctx, s3Conf)
	if err != nil {
		log.Fatalf("Could not create S3 filestore: %v", err)
	}

	s3Bucket := os.Getenv("S3_BUCKET_NAME")
	if s3Bucket == "" {
		log.Fatal("S3_BUCKET_NAME is not set")
	}

	log.Println("S3 FileStore initialized")

	router := api.NewRouter(dbStore.Pool, redisClient.Client, fileStore, s3Bucket)

	http.ListenAndServe("8080", router)
}
