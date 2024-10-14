package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

type courseGetter interface {
	GetCourses(ctx context.Context) ([]models.Course, error)
	GetCourse(ctx context.Context, id int) (models.Course, error)
	CreateCourse(ctx context.Context, course models.Course) (models.Course, error)
	UpdateCourse(ctx context.Context, id int, course models.Course) (models.Course, error)
	DeleteCourse(ctx context.Context, id int) error
}

func HandleGetCourses(logger *httplog.Logger, service courseGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		courses, err := service.GetCourses(ctx)
		if err != nil {
			logger.Error("error getting all courses", "error", err)
			EncodeResponse(w, logger, http.StatusInternalServerError, ResponseErr{Error: "Error retrieving data"})
			return
		}
		EncodeResponse(w, logger, http.StatusOK, courses)
	}
}

func HandleGetCourse(logger *httplog.Logger, service courseGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idParam := chi.URLParam(r, "id")

		id, err := strconv.Atoi(idParam)
		if err != nil {
			logger.Error("invalid course ID", "error", err)
			EncodeResponse(w, logger, http.StatusBadRequest, ResponseErr{Error: "Invalid course ID"})
			return
		}

		course, err := service.GetCourse(ctx, id)
		if err != nil {
			logger.Error("error getting course", "error", err)
			EncodeResponse(w, logger, http.StatusInternalServerError, ResponseErr{Error: "Error retrieving data"})
			return
		}
		EncodeResponse(w, logger, http.StatusOK, course)
	}
}

func HandleCreateCourse(logger *httplog.Logger, service courseGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var course models.Course
		if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
			logger.Error("failed to decode request body", "error", err)
			EncodeResponse(w, logger, http.StatusBadRequest, ResponseErr{Error: "Invalid request payload"})
			return
		}
		if course.Name == "" {
			logger.Error("course name is missing", "error", fmt.Errorf("missing course name"))
			EncodeResponse(w, logger, http.StatusBadRequest, ResponseErr{Error: "Course name is required"})
			return
		}
		course, err := service.CreateCourse(ctx, course)
		if err != nil {
			logger.Error("error creating course", "error", err)
			EncodeResponse(w, logger, http.StatusInternalServerError, ResponseErr{Error: "Error creating data"})
			return
		}
		EncodeResponse(w, logger, http.StatusOK, course.ID)
	}
}

func HandleUpdateCourse(logger *httplog.Logger, service courseGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var course models.Course
		idParam := chi.URLParam(r, "id")

		id, err := strconv.Atoi(idParam)
		if err != nil {
			logger.Error("invalid course ID", "error", err)
			EncodeResponse(w, logger, http.StatusBadRequest, ResponseErr{Error: "Invalid course ID"})
			return
		}

		err = json.NewDecoder(r.Body).Decode(&course)
		if err != nil {
			logger.Error("failed to decode request body", "error", err)
			EncodeResponse(w, logger, http.StatusBadRequest, ResponseErr{Error: "Invalid request payload"})
			return
		}
		course, err = service.UpdateCourse(ctx, id, course)
		if err != nil {
			logger.Error("error updating course", "error", err)
			EncodeResponse(w, logger, http.StatusInternalServerError, ResponseErr{Error: "Error updating data"})
			return
		}
		EncodeResponse(w, logger, http.StatusOK, course)
	}
}

func HandleDeleteCourse(logger *httplog.Logger, service courseGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		idParam := chi.URLParam(r, "id")

		id, err := strconv.Atoi(idParam)
		if err != nil {
			logger.Error("invalid course ID", "error", err)
			EncodeResponse(w, logger, http.StatusBadRequest, ResponseErr{Error: "Invalid course ID"})
			return
		}

		if err := service.DeleteCourse(ctx, id); err != nil {
			logger.Error("error deleting course", "error", err)
			EncodeResponse(w, logger, http.StatusInternalServerError, ResponseErr{Error: "Error deleting data"})
			return
		}
		EncodeResponse(w, logger, http.StatusOK, "Course has successfully been deleted")
	}
}
