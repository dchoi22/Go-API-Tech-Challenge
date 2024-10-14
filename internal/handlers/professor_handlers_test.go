package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/handlers"
	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockProfessorGetter struct {
	mock.Mock
}

func (m *mockProfessorGetter) GetPeople(ctx context.Context, firstName, lastName, age, personType string) ([]models.Person, error) {
	args := m.Called(ctx, firstName, lastName, age, personType)
	return args.Get(0).([]models.Person), args.Error(1)
}

func (m *mockProfessorGetter) GetPerson(ctx context.Context, firstName, personType string) (models.Person, error) {
	args := m.Called(ctx, firstName, personType)
	return args.Get(0).(models.Person), args.Error(1)
}

func (m *mockProfessorGetter) UpdatePerson(ctx context.Context, firstName, personType string, person models.Person) (models.Person, error) {
	args := m.Called(ctx, firstName, personType, person)
	return args.Get(0).(models.Person), args.Error(1)
}

func (m *mockProfessorGetter) CreatePerson(ctx context.Context, person models.Person) (models.Person, error) {
	args := m.Called(ctx, person)
	return args.Get(0).(models.Person), args.Error(1)
}

func (m *mockProfessorGetter) DeletePerson(ctx context.Context, firstName, personType string) error {
	args := m.Called(ctx, firstName, personType)
	return args.Error(0)
}

func (m *mockProfessorGetter) UpdatePersonCourses(ctx context.Context, studentID int, newCourses []int64) error {
	return nil
}

func TestHandleGetProfessors(t *testing.T) {
	tests := []struct {
		name           string
		firstName      string
		lastName       string
		age            string
		mockPeople     []models.Person
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Success",
			firstName:      "John",
			lastName:       "Doe",
			age:            "40",
			mockPeople:     []models.Person{{FirstName: "John", LastName: "Doe", Type: "professor"}},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   []models.Person{{FirstName: "John", LastName: "Doe", Type: "professor"}},
		},
		{
			name:           "Error Fetching Professors",
			firstName:      "error",
			lastName:       "Doe",
			age:            "40",
			mockPeople:     nil,
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handlers.ResponseErr{Error: "Error retrieving data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockProfessorGetter)

			// Mock GetPeople based on test scenario
			mockService.On("GetPeople", mock.Anything, tt.firstName, tt.lastName, tt.age, "professor").Return(tt.mockPeople, tt.mockError)

			logger := httplog.NewLogger("test", httplog.Options{})
			handler := handlers.HandleGetProfessors(logger, mockService)

			req, _ := http.NewRequest("GET", "/professors?first-name="+tt.firstName+"&last-name="+tt.lastName+"&age="+tt.age, nil)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/professors", handler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var people []models.Person
				err := json.Unmarshal(rr.Body.Bytes(), &people)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockPeople, people)
			} else {
				var errorResponse handlers.ResponseErr
				err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(handlers.ResponseErr).Error, errorResponse.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleGetProfessor(t *testing.T) {
	tests := []struct {
		name           string
		firstName      string
		mockPerson     models.Person
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Success",
			firstName:      "John",
			mockPerson:     models.Person{FirstName: "John", LastName: "Doe", Type: "professor"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   models.Person{FirstName: "John", LastName: "Doe", Type: "professor"},
		},
		{
			name:           "Error Fetching Professor",
			firstName:      "error",
			mockPerson:     models.Person{},
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handlers.ResponseErr{Error: "Error retrieving data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockProfessorGetter)

			// Mock GetPerson based on test scenario
			mockService.On("GetPerson", mock.Anything, tt.firstName, "professor").Return(tt.mockPerson, tt.mockError)

			logger := httplog.NewLogger("test", httplog.Options{})
			handler := handlers.HandleGetProfessor(logger, mockService)

			req, _ := http.NewRequest("GET", "/professors/"+tt.firstName, nil)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/professors/{firstName}", handler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var professor models.Person
				err := json.Unmarshal(rr.Body.Bytes(), &professor)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockPerson, professor)
			} else {
				var errorResponse handlers.ResponseErr
				err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(handlers.ResponseErr).Error, errorResponse.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleDeleteProfessor(t *testing.T) {
	tests := []struct {
		name           string
		firstName      string
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Success",
			firstName:      "John",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `"Professor has successfully been deleted"`,
		},
		{
			name:           "Error Deleting Professor",
			firstName:      "John",
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handlers.ResponseErr{Error: "Error deleting data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockProfessorGetter)

			// Only mock DeletePerson when the request is expected to be valid
			if tt.expectedStatus != http.StatusBadRequest {
				mockService.On("DeletePerson", mock.Anything, tt.firstName, "professor").Return(tt.mockError)
			}

			logger := httplog.NewLogger("test", httplog.Options{})
			handler := handlers.HandleDeleteProfessor(logger, mockService)

			req, _ := http.NewRequest("DELETE", "/professors/"+tt.firstName, nil)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Delete("/professors/{firstName}", handler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody, strings.TrimSpace(rr.Body.String()))
			} else {
				var errorResponse handlers.ResponseErr
				err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(handlers.ResponseErr).Error, errorResponse.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}
