package api

import (
	"log"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/google/uuid"
	"job-matcher/internal/objectstore"
	"net/http"
	"path/filepath"
	"job-matcher/internal/storage"
	"job-matcher/internal/queue"
)

// TODO
// Add s3
// Add job to postgres with status "queued"
// get jobid back from postgres
// add job to redis

type APIHandler struct {
	db       storage.JobStore
	queue    queue.JobQueuer
	store    objectstore.FileStorer
	s3Bucket string
}


func NewAPIHandler(db storage.JobStore, queue queue.JobQueuer, store objectstore.FileStorer, s3Bucket string) *APIHandler { 	
	return &APIHandler{
	db: db,
	queue: queue,
	store: store, 
	s3Bucket: s3Bucket,
	}

}

func (h *APIHandler) HandleUploadResume(w http.ResponseWriter, r *http.Request) {

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
	fileURL := output.Location
	
	if fileURL == "" { 
		http.Error(w, "The file location was not successfully retrieved", http.StatusInternalServerError)
		return 	
	}


	jobID, err := h.db.InsertJobAndGetID(r.Context(), fileURL)	
	
	if !jobID.Valid {
		http.Error(w, "Unable to find job status", http.StatusInternalServerError)
		return 
	}


	if err != nil {
    	http.Error(w, "Failed to insert job into database", http.StatusInternalServerError)
    	return
	}
	if !jobID.Valid {
    http.Error(w, "Failed to create a valid job ID", http.StatusInternalServerError)
    return
	}

	jobIDString := jobID.UUID.String()
	

	err = h.queue.InsertJob(r.Context(), jobIDString)

	if err != nil { 	
		http.Error(w, "Failed to insert job into the worker pool", http.StatusInternalServerError)
		return 
	}


		
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	log.Printf("Job %s queued successfully at %s", jobIDString, fileURL)
	json.NewEncoder(w).Encode(map[string]string{"jobId": jobIDString})
}
