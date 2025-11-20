package api

import (
	"net/http"
)

func NewRouter(h *APIHandler) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("POST /resumes", h.HandleUploadResume)

	mux.HandleFunc("GET /resumes/{jobId}", h.HandleViewResult)

	return mux
}
