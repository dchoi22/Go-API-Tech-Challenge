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

type professorGetter interface {
	GetPeople(ctx context.Context, firstName, lastName, age, personType string) ([]models.Person, error)
	GetPerson(ctx context.Context, firstName, personType string) (models.Person, error)
	UpdatePerson(ctx context.Context, firstName, personType string, person models.Person) (models.Person, error)
	CreatePerson(ctx context.Context, person models.Person) (models.Person, error)
	DeletePerson(ctx context.Context, firstName, personType string) error
}

func HandleGetProfessors(logger *httplog.Logger, service professorGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queryParams := r.URL.Query()
		firstName := queryParams.Get("first-name")
		lastName := queryParams.Get("last-name")
		age := queryParams.Get("age")

		professors, err := service.GetPeople(ctx, firstName, lastName, age, "professor")
		if err != nil {
			logger.Error("error getting all professors", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error retrieving data"})
			return
		}
		encodeResponse(w, logger, http.StatusOK, professors)
	}
}

func HandleGetProfessor(logger *httplog.Logger, service professorGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		nameParam := chi.URLParam(r, "firstName")

		professor, err := service.GetPerson(ctx, nameParam, "professor")
		if err != nil {
			logger.Error("error getting professor", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error retrieving data"})
			return
		}
		encodeResponse(w, logger, http.StatusOK, professor)
	}
}

func HandleUpdateProfessor(logger *httplog.Logger, service professorGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		nameParam := chi.URLParam(r, "firstName")
		var professor models.Person

		err := json.NewDecoder(r.Body).Decode(&professor)
		if err != nil {
			logger.Error("failed to decode request body", "error", err)
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: "Invalid request payload"})
			return
		}

		if err := utils.ValidatePerson(professor); err != nil {
			logger.Error("invalid professor data", "error", err)
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: err.Error()})
			return
		}

		professor, err = service.UpdatePerson(ctx, nameParam, "professor", professor)
		if err != nil {
			logger.Error("error updating professor", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error updating data"})
			return
		}
		encodeResponse(w, logger, http.StatusOK, professor)
	}
}

func HandleCreateProfessor(logger *httplog.Logger, service professorGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var professor models.Person
		if err := json.NewDecoder(r.Body).Decode(&professor); err != nil {
			logger.Error("failed to decode request body", "error", err)
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: "Invalid request payload"})
			return
		}
		if err := utils.ValidatePerson(professor); err != nil {
			logger.Error("invalid professor data", "error", err)
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: err.Error()})
			return
		}
		professor, err := service.CreatePerson(ctx, professor)
		if err != nil {
			logger.Error("error creating professor", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error creating data"})
			return
		}
		encodeResponse(w, logger, http.StatusOK, professor.ID)
	}
}

func HandleDeleteProfessor(logger *httplog.Logger, service professorGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		nameParam := chi.URLParam(r, "firstName")
		if nameParam == "" {
			logger.Error("invalid professor name", "error", errors.New("first name is required"))
			encodeResponse(w, logger, http.StatusBadRequest, responseErr{Error: "Invalid professor name"})
			return
		}
		if err := service.DeletePerson(ctx, nameParam, "professor"); err != nil {
			logger.Error("error deleting professor", "error", err)
			encodeResponse(w, logger, http.StatusInternalServerError, responseErr{Error: "Error deleting data"})
			return
		}
		encodeResponse(w, logger, http.StatusOK, "Professor has successfully been deleted")
	}
}
