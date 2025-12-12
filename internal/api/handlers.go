package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"job-matcher/internal/models"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type APIHandler struct {
	job      JobStore
	queue    Producer
	uploader Uploader
	s3Bucket string
}

type JobStore interface {
	Create(ctx context.Context, job *models.Job) error
	Get(ctx context.Context, jobID uuid.UUID) (*models.Job, error)
}

type Producer interface {
	Produce(ctx context.Context, jobID uuid.UUID, s3Key string) error
}

type Uploader interface {
	Upload(ctx context.Context, file io.Reader, bucket, key, contentType string) error
}

func NewAPIHandler(db JobStore, queue Producer, store Uploader, s3Bucket string) *APIHandler {
	return &APIHandler{
		job:      db,
		queue:    queue,
		uploader: store,
		s3Bucket: s3Bucket,
	}

}

func (h *APIHandler) UploadResume(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	file, fileHeader, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, "An error occurred upon retrieving the file.", http.StatusBadRequest)
		return
	}
	defer file.Close() // upload file

	newJobID, _ := uuid.NewV7()

	uniqueFileName := fmt.Sprintf("%s-%s", newJobID.String(), filepath.Ext(fileHeader.Filename))

	err = h.uploader.Upload(r.Context(), file, h.s3Bucket, uniqueFileName, "application/pdf")

	if err != nil {
		http.Error(w, "Failed to upload file ", http.StatusInternalServerError)
		return
	}

	newJob := &models.Job{
        ID:        newJobID,
        FileName:  uniqueFileName,
        Status: models.StatusQueued, 
        CreatedAt: time.Now(),             
    }

	h.job.Create(r.Context(), newJob)

	if err != nil {
		http.Error(w, "An error occurred while processing your resume", http.StatusInternalServerError)
		return
	}

	err = h.queue.Produce(r.Context(), newJobID, uniqueFileName)

	if err != nil {
		http.Error(w, "An error occurred while processing your resume", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	log.Printf("Job %s queued successfully at %s", newJobID.String(), uniqueFileName)
	if err := json.NewEncoder(w).Encode(newJobID); err != nil { 
		log.Printf("Failed to encode response: %v", err)
	}
}

func (h *APIHandler) ViewResult(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	jobIDString := r.URL.Query().Get("id")

	if jobIDString == "" {
		http.Error(w, "Missing job ID. Please provide it as a query parameter: ?id=UUID", http.StatusBadRequest)
		return
	}

	jobID, err := uuid.Parse(jobIDString)

	if err != nil {
		http.Error(w, "Invalid job id format", http.StatusBadRequest)
	}

	jobData, err := h.job.Get(r.Context(), jobID)

	if err != nil {
		log.Printf("Error retrieving job %s: %v", jobIDString, err)
		http.Error(w, "Job not found or database error.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(jobData); err != nil { 
		log.Printf("Failed to encode response: %v", err)
	}
}

