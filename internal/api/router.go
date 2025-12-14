package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func NewRouter(h *APIHandler) http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestSize(5 * 1024 * 1024))

	r.Post("/upload", h.UploadResume)

	r.Get("/resumes/{jobID}", h.ViewResult)

	return r
}
