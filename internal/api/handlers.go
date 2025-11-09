package api

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"job-matcher/internal/objectstore"
	"job-matcher/internal/queue"
	"job-matcher/internal/storage"
	"log"
	"net/http"
	"path/filepath"
)


type APIHandler struct {
	db       storage.JobCreator
	queue    queue.JobProducer
	store    objectstore.FileStorer
	s3Bucket string
}

func NewAPIHandler(db storage.JobCreator, queue queue.JobProducer, store objectstore.FileStorer, s3Bucket string) *APIHandler {
	return &APIHandler{
		db:       db,
		queue:    queue,
		store:    store,
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
		http.Error(w, "An error occurred upon retrieving the file.", http.StatusBadRequest)
		return
	}
	defer file.Close() // upload file

	uniqueFileName := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(fileHeader.Filename))

	err = h.store.Upload(r.Context(), file, h.s3Bucket, uniqueFileName, "application/pdf")

	if err != nil {
		http.Error(w, "Failed to upload file ", http.StatusInternalServerError)
		return
	}

	jobID, err := h.db.InsertJobReturnID(r.Context(), uniqueFileName, storage.JobStatus(storage.Queued))

	if err != nil {
		http.Error(w, "An error occurred while processing your resume", http.StatusInternalServerError)
		return
	}
	if jobID == uuid.Nil {
		http.Error(w, "An error occurred while processing your resume", http.StatusInternalServerError)
		return
	}

	jobIDString := jobID.String()

	err = h.queue.InsertJob(r.Context(), jobIDString)

	if err != nil {
		http.Error(w, "An error occurred while processing your resume", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	log.Printf("Job %s queued successfully at %s", jobIDString, uniqueFileName)
	json.NewEncoder(w).Encode(map[string]string{"jobId": jobIDString})
}
