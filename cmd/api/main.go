package main

import (
	"context"
	"job-matcher/internal/api"
	"job-matcher/internal/postgresdb"
	"job-matcher/internal/s3"
	"job-matcher/internal/valkeydb"
	"log"
	"net/http"
	"os"
)

func main() {

	ctx := context.Background()
	postgresDB, err := postgresdb.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		panic(err)
	}
	defer postgresDB.Close()

	valkeyQueue, err := valkeydb.New(ctx, os.Getenv("VALKEY_URL"), os.Getenv("VALKEY_PASSWORD"))

	if err != nil {
		panic(err)
	}

	defer valkeyQueue.Close()

	s3Conf := s3.S3Config{
		EndpointURL: os.Getenv("S3_ENDPOINT_URL"),
		Region:      os.Getenv("S3_REGION"),
		AccessKey:   os.Getenv("S3_ACCESS_KEY"),
		SecretKey:   os.Getenv("S3_SECRET_KEY"),
	}

	// 2. Create the FileStore
	s3Store, err := s3.NewFileStore(ctx, s3Conf)

	if err != nil {
		log.Fatalf("Could not create S3 filestore: %v", err)
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("S3_BUCKET_NAME is not set")
	}

	log.Println("S3 FileStore initialized")

	apiHandler := api.NewAPIHandler(postgresDB, valkeyQueue, s3Store, bucketName)

	router := api.NewRouter(apiHandler)

	http.ListenAndServe("8080", router)
}
