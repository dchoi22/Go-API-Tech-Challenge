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

type mockStudentGetter struct {
	mock.Mock
}

func (m *mockStudentGetter) GetPeople(ctx context.Context, firstName, lastName, age, personType string) ([]models.Person, error) {
	args := m.Called(ctx, firstName, lastName, age, personType)
	return args.Get(0).([]models.Person), args.Error(1)
}

func (m *mockStudentGetter) GetPerson(ctx context.Context, firstName, personType string) (models.Person, error) {
	args := m.Called(ctx, firstName, personType)
	return args.Get(0).(models.Person), args.Error(1)
}

func (m *mockStudentGetter) UpdatePerson(ctx context.Context, firstName, personType string, person models.Person) (models.Person, error) {
	args := m.Called(ctx, firstName, personType, person)
	return args.Get(0).(models.Person), args.Error(1)
}

func (m *mockStudentGetter) CreatePerson(ctx context.Context, person models.Person) (models.Person, error) {
	args := m.Called(ctx, person)
	return args.Get(0).(models.Person), args.Error(1)
}

func (m *mockStudentGetter) DeletePerson(ctx context.Context, firstName, personType string) error {
	args := m.Called(ctx, firstName, personType)
	return args.Error(0)
}

func (m *mockStudentGetter) UpdatePersonCourses(ctx context.Context, studentID int, newCourses []int64) error {
	return nil
}

func TestHandleDeleteStudent(t *testing.T) {
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
			expectedBody:   `"Student has successfully been deleted"`,
		},
		{
			name:           "Error Deleting Student",
			firstName:      "John",
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handlers.ResponseErr{Error: "Error deleting data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockStudentGetter)

			// Only mock DeletePerson when the request is expected to be valid

			mockService.On("DeletePerson", mock.Anything, tt.firstName, "student").Return(tt.mockError)

			logger := httplog.NewLogger("test")
			handler := handlers.HandleDeleteStudent(logger, mockService)

			req, _ := http.NewRequest("DELETE", "/api/student/"+tt.firstName, nil)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Delete("/api/student/{firstName}", handler)
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

func TestHandleGetStudents(t *testing.T) {
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
			age:            "20",
			mockPeople:     []models.Person{{FirstName: "John", LastName: "Doe", Type: "student"}},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   []models.Person{{FirstName: "John", LastName: "Doe", Type: "student"}},
		},
		{
			name:           "Error Fetching Students",
			firstName:      "error",
			lastName:       "Doe",
			age:            "20",
			mockPeople:     nil,
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handlers.ResponseErr{Error: "Error retrieving data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockStudentGetter)

			// Mock GetPeople based on test scenario
			mockService.On("GetPeople", mock.Anything, tt.firstName, tt.lastName, tt.age, "student").Return(tt.mockPeople, tt.mockError)

			logger := httplog.NewLogger("test", httplog.Options{})
			handler := handlers.HandleGetStudents(logger, mockService)

			req, _ := http.NewRequest("GET", "/students?first-name="+tt.firstName+"&last-name="+tt.lastName+"&age="+tt.age, nil)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/students", handler)
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

func TestHandleGetStudent(t *testing.T) {
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
			mockPerson:     models.Person{FirstName: "John", LastName: "Doe", Type: "student"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   models.Person{FirstName: "John", LastName: "Doe", Type: "student"},
		},
		{
			name:           "Error Fetching Student",
			firstName:      "error",
			mockPerson:     models.Person{},
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handlers.ResponseErr{Error: "Error retrieving data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockStudentGetter)

			// Mock GetPerson based on test scenario
			mockService.On("GetPerson", mock.Anything, tt.firstName, "student").Return(tt.mockPerson, tt.mockError)

			logger := httplog.NewLogger("test", httplog.Options{})
			handler := handlers.HandleGetStudent(logger, mockService)

			req, _ := http.NewRequest("GET", "/students/"+tt.firstName, nil)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/students/{firstName}", handler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var student models.Person
				err := json.Unmarshal(rr.Body.Bytes(), &student)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockPerson, student)
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
