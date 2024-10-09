package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/handlers/utils"
	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

type studentGetter interface {
	GetPeople(ctx context.Context, firstName, lastName, age, personType string) ([]models.Person, error)
	GetPerson(ctx context.Context, firstName, personType string) (models.Person, error)
	UpdatePerson(ctx context.Context, firstName, personType string, person models.Person) (models.Person, error)
	CreatePerson(ctx context.Context, person models.Person) (models.Person, error)
	DeletePerson(ctx context.Context, firstName, personType string) error
	UpdatePersonCourses(ctx context.Context, studentID int, newCourses []int64) error
}

func HandleGetStudents(logger *httplog.Logger, service studentGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queryParams := r.URL.Query()
		firstName := queryParams.Get("first-name")
		lastName := queryParams.Get("last-name")
		age := queryParams.Get("age")

		students, err := service.GetPeople(ctx, firstName, lastName, age, "student")
		if err != nil {
			logger.Error("error getting all students", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error retrieving data"})
			return
		}
		encodeResponse(w, logger, http.StatusOK, students)
	}
}

func HandleGetStudent(logger *httplog.Logger, service studentGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		nameParam := chi.URLParam(r, "firstName")

		student, err := service.GetPerson(ctx, nameParam, "student")
		if err != nil {
			logger.Error("error getting student", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error retrieving data"})
			return
		}
		encodeResponse(w, logger, http.StatusOK, student)
	}
}

func HandleUpdateStudent(logger *httplog.Logger, service studentGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		nameParam := chi.URLParam(r, "firstName")
		var student models.Person

		// Decode request body
		err := json.NewDecoder(r.Body).Decode(&student)
		if err != nil {
			logger.Error("failed to decode request body", "error", err)
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: "Invalid request payload"})
			return
		}

		// Validate the student data
		if err := utils.ValidatePerson(student); err != nil {
			logger.Error("invalid student data", "error", err)
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: err.Error()})
			return
		}

		if student.Type != "student" {
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: "Person is not of type student"})
			return
		}

		// Get the existing student to retrieve the ID
		existingStudent, err := service.GetPerson(ctx, nameParam, "student")
		if err != nil {
			logger.Error("error fetching existing student", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error updating data"})
			return
		}

		// Set the ID from the existing student
		student.ID = existingStudent.ID

		// Update the student in the database
		updatedStudent, err := service.UpdatePerson(ctx, nameParam, "student", student)
		if err != nil {
			logger.Error("error updating student", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error updating data"})
			return
		}

		err = service.UpdatePersonCourses(ctx, student.ID, updatedStudent.Courses)
		if err != nil {
			logger.Error("error updating student's courses", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error updating courses"})
			return
		}

		// Respond with the updated student object
		encodeResponse(w, logger, http.StatusOK, updatedStudent)
	}
}

func HandleCreateStudent(logger *httplog.Logger, service studentGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var student models.Person
		if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
			logger.Error("failed to decode request body", "error", err)
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: "Invalid request payload"})
			return
		}
		if err := utils.ValidatePerson(student); err != nil {
			logger.Error("invalid student data", "error", err)
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: err.Error()})
			return
		}

		if student.Type != "student" {
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: "Person is not of type student"})
			return
		}
		student, err := service.CreatePerson(ctx, student)
		if err != nil {
			logger.Error("error creating student", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error creating data"})
			return
		}

		if len(student.Courses) > 0 {
			err = service.UpdatePersonCourses(ctx, student.ID, student.Courses)
			if err != nil {
				logger.Error("error associating courses with student", "error", err)
				encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error associating courses"})
				return
			}
		}
		encodeResponse(w, logger, http.StatusOK, student.ID)
	}
}

func HandleDeleteStudent(logger *httplog.Logger, service studentGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		nameParam := chi.URLParam(r, "firstName")
		if nameParam == "" {
			logger.Error("invalid student name", "error", errors.New("first name is required"))
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: "Invalid student name"})
			return
		}
		if err := service.DeletePerson(ctx, nameParam, "student"); err != nil {
			logger.Error("error deleting student", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error deleting data"})
			return
		}
		encodeResponse(w, logger, http.StatusOK, "Student has successfully been deleted")
	}
}
