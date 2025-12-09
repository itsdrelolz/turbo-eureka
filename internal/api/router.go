package api

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *APIHandler) http.Handler {


	r := chi.NewRouter() 


	r.Use(middleware.RequestID)
  	r.Use(middleware.RealIP)
  	r.Use(middleware.Logger)
  	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestSize( 5 * 1024 * 1024))

	r.Get("/health", h.CheckHealth)
	r.Post("/resumes", h.UploadResume)

	r.Get("/resumes/{jobID}", h.ViewResult)

	r.Get("/resumes/{jobID}/status", h.TODO)
	return r
}

