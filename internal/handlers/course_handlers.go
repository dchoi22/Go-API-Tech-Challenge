package handlers

import (
	"context"
	"net/http"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

type courseGetter interface {
	GetCourses(ctx context.Context) ([]models.Course, error)
	GetCourse(ctx context.Context, id string) (models.Course, error)
	CreateCourse(ctx context.Context) (models.Course, error)
	UpdateCourse(ctx context.Context, id string) (models.Course, error)
	DeleteCourse(ctx context.Context, id string) (models.Course, error)
}

func HandleGetCourses(logger *httplog.Logger, service courseGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		courses, err := service.GetCourses(ctx)
		if err != nil {
			logger.Error("error getting all courses", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error retrieving data"})
			return
		}
		encodeResponse(w, logger, http.StatusOK, courses)
	}
}

func HandleGetCourse(logger *httplog.Logger, service courseGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")

		course, err := service.GetCourse(ctx, id)
		if err != nil {
			logger.Error("error getting course", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error retrieving data"})
			return
		}
		encodeResponse(w, logger, http.StatusOK, course)
	}
}


