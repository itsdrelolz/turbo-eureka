package api

import (
	"job-matcher/internal/objectstore"
	"net/http"	
	"github.com/redis/go-redis/v9"
	"job-matcher/internal/database"
)


func NewRouter(db *database.Store, queue *redis.Client, fs *objectstore.FileStore, bucketName string) http.Handler {

	h := &handler{
		db:       db,
		redis:    queue,
		store:    fs,
		s3Bucket: bucketName,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /resumes", h.handleUploadResume)

	return mux
}
