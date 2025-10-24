package api

import (
	"job-matcher/internal/objectstore"
	"net/http"
	"job-matcher/internal/database"
	"job-matcher/internal/queue"
)


func NewRouter(db *database.Store, queue *queue.ValkeyClient, fs *objectstore.FileStore, bucketName string) http.Handler {

	h := &handler{
		db:       db,
		queue:    queue,
		store:    fs,
		s3Bucket: bucketName,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /resumes", h.handleUploadResume)

	return mux
}
