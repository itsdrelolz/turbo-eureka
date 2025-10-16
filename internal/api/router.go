package api

import (
	"net/http"
)



func newRouter(db *pgxpool.Pool, queue *redis.Client, s3Client *s3.Client, bucketName string) http.Handler { 

	h := &hander{ 
		db: db, 
		queue: queue, 
		s3: s3Client, 
		s3Bucket: bucketName,
	}


	mux := http.NewServeMux() 

	mux.HandleFunc("POST /resumes", h.handleUploadResume)

	return mux 
}
