package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"job-matcher/internal/objectstore"
	"net/http"
	"path/filepath"
	"job-matcher/internal/database"	
)

// TODO
// Add s3
// Add job to postgres with status "queued"
// get jobid back from postgres
// add job to redis

type handler struct {
	db       *database.Store
	redis    *redis.Client
	store    *objectstore.FileStore
	s3Bucket string
}

func (h *handler) handleUploadResume(w http.ResponseWriter, r *http.Request) {

	const MAX_UPLOAD_SIZE = 5 << 20 // 5 MB

	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)

	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, "The uploaded file is too large. Please choose a file smaller than 5MB.", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, "Invalid file key.", http.StatusBadRequest)
		return
	}
	defer file.Close() // upload file

	uniqueFileName := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(fileHeader.Filename))

	output, err := h.store.Upload(r.Context(), file, h.s3Bucket, uniqueFileName, "application/pdf")
	if err != nil {

		var mu manager.MultiUploadFailure
		if errors.As(err, &mu) {
			// Process error and its associated uploadID
			fmt.Println("Error:", mu)
			_ = mu.UploadID() // retrieve the associated UploadID
		} else {
			// Process error generically
			fmt.Println("Error:", err.Error())
		}
		http.Error(w, "Failed to upload file.", http.StatusInternalServerError)
		return
	}

	// TODO
	// This section shoulw get the file location from s3, and then insert it into the postgresdb as a job, with the status 'queued' .
	// it then gets the jobid back and adds this job to the redis queue for workers to handle
	// finally it should return to the user the jobid so they can check its status
	fileURL := output.Location
	
	if fileURL == "" { 
		http.Error(w, "The file location was not successfully retrieved", http.StatusInternalServerError)
		return 	
	}


	jobID, err := h.db.InsertJobAndGetID(r.Context(), fileURL)	
	
	if jobID == "" {
		http.Error(w, "Unable to find job status", http.StatusInternalServerError)
		return 
	}

	// TODO: 
	// Now add this jobid to the redis queue 
		
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"jobId": jobID})
}
