package main

import (
	"context"
	"github.com/joho/godotenv"
	"job-matcher/internal/api"
	"job-matcher/internal/postgres"
	"job-matcher/internal/redis"
	"job-matcher/internal/s3"
	"log"
	"log/slog"
	"net/http"
	"os"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	ctx := context.Background()

	db, err := postgresdb.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatalln("Failed to initialize postgres")
	}

	defer db.Close()

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
		log.Fatalln("Could not create S3 store")
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")

	if bucketName == "" {
		log.Fatalln("S3_BUCKET_NAME is not set")
	}

	log.Println("S3 FileStore initialized")

	apiHandler := api.NewAPIHandler(db, redis, s3Store, bucketName)

	router := api.NewRouter(apiHandler)

	http.ListenAndServe(":8080", router)
}
