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

	"github.com/google/uuid"
)

type APIHandler struct {
	job      JobStore
	queue    Producer
	uploader Uploader
	s3Bucket string
}

type JobStore interface {
	Create(ctx context.Context, jobID uuid.UUID, fileName string)
	GetResult(ctx context.Context, jobID uuid.UUID) (*models.Job, error)
}

type Producer interface {
	Produce(ctx context.Context, jobID uuid.UUID) error
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

func (h *APIHandler) HandleUploadResume(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	const MAX_UPLOAD_SIZE = 5 << 20 // 5 MB

	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)

	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, "The uploaded file is too large. Please choose a file smaller than 5MB.", http.StatusBadRequest)
		return
	}

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

	h.job.Create(r.Context(), newJobID, uniqueFileName)

	if err != nil {
		http.Error(w, "An error occurred while processing your resume", http.StatusInternalServerError)
		return
	}

	err = h.queue.Produce(r.Context(), newJobID)

	if err != nil {
		http.Error(w, "An error occurred while processing your resume", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	log.Printf("Job %s queued successfully at %s", newJobID.String(), uniqueFileName)
	json.NewEncoder(w).Encode(map[string]string{"job_id": newJobID.String()})
}

func (h *APIHandler) HandleViewResult(w http.ResponseWriter, r *http.Request) {

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

	jobData, err := h.job.GetResult(r.Context(), jobID)

	if err != nil {
		log.Printf("Error retrieving job %s: %v", jobIDString, err)
		http.Error(w, "Job not found or database error.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"job_id": jobIDString,
		"status":      jobData.JobStatus.String(),
		"resume_text": jobData.ResumeText,
	})
}
