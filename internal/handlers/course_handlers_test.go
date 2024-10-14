package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/handlers"
	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCourseGetter struct {
	mock.Mock
}

func (m *mockCourseGetter) GetCourses(ctx context.Context) ([]models.Course, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *mockCourseGetter) GetCourse(ctx context.Context, id int) (models.Course, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Course), args.Error(1)
}

func (m *mockCourseGetter) CreateCourse(ctx context.Context, course models.Course) (models.Course, error) {
	args := m.Called(ctx, course)
	return args.Get(0).(models.Course), args.Error(1)
}

func (m *mockCourseGetter) UpdateCourse(ctx context.Context, id int, course models.Course) (models.Course, error) {
	args := m.Called(ctx, id, course)
	return args.Get(0).(models.Course), args.Error(1)
}
func (m *mockCourseGetter) DeleteCourse(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestHandleGetCourses(t *testing.T) {
	tests := []struct {
		name           string
		mockCourses    []models.Course
		mockError      error
		expectedStatus int
		expectedBody   []models.Course
	}{
		{
			name:           "Success",
			mockCourses:    []models.Course{{ID: 1, Name: "Test Course"}, {ID: 2, Name: "Test 2 Course"}},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   []models.Course{{ID: 1, Name: "Test Course"}, {ID: 2, Name: "Test 2 Course"}},
		},
		{
			name:           "Error",
			mockCourses:    nil,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockCourseGetter)
			mockService.On("GetCourses", mock.Anything).Return(tt.mockCourses, tt.mockError)

			logger := httplog.NewLogger("test")

			handler := handlers.HandleGetCourses(logger, mockService)

			req, _ := http.NewRequest("GET", "/api/course", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []models.Course
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			} else {
				var errorResponse handlers.ResponseErr
				err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, "Error retrieving data", errorResponse.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleGetCourse(t *testing.T) {
	tests := []struct {
		name           string
		courseID       string
		mockCourses    models.Course
		mockError      error
		expectedStatus int
		expectedBody   models.Course
	}{
		{
			name:           "Success",
			courseID:       "1",
			mockCourses:    models.Course{ID: 1, Name: "Test Course"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   models.Course{ID: 1, Name: "Test Course"},
		},
		{
			name:           "Error",
			courseID:       "1",
			mockCourses:    models.Course{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   models.Course{},
		},
		{
			name:           "Invalid ID",
			courseID:       "abc",
			mockCourses:    models.Course{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   models.Course{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockCourseGetter)

			if tt.expectedStatus != http.StatusBadRequest {
				courseIDInt, _ := strconv.Atoi(tt.courseID)
				mockService.On("GetCourse", mock.Anything, courseIDInt).Return(tt.mockCourses, tt.mockError)
			}

			logger := httplog.NewLogger("test")

			handler := handlers.HandleGetCourse(logger, mockService)

			req, _ := http.NewRequest("GET", "/api/course/"+tt.courseID, nil)

			r := chi.NewRouter()
			r.Get("/api/course/{id}", handler)

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.Course
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			} else if tt.expectedStatus == http.StatusInternalServerError {
				var errorResponse handlers.ResponseErr
				err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, "Error retrieving data", errorResponse.Error)
			} else if tt.expectedStatus == http.StatusBadRequest {
				var errorResponse handlers.ResponseErr
				err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, "Invalid course ID", errorResponse.Error)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleCreateCourse(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		inputCourse    models.Course
		returnedCourse models.Course
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Success",
			requestBody:    `{"Name": "Test Course"}`,
			inputCourse:    models.Course{Name: "Test Course"},
			returnedCourse: models.Course{ID: 1, Name: "Test Course"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   1,
		},
		{
			name:           "Invalid Request Body",
			requestBody:    `{}`,
			inputCourse:    models.Course{},
			returnedCourse: models.Course{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   handlers.ResponseErr{Error: "Course name is required"},
		},
		{
			name:           "Error Creating Course",
			requestBody:    `{"Name": "Test Course"}`,
			inputCourse:    models.Course{Name: "Test Course"},
			returnedCourse: models.Course{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handlers.ResponseErr{Error: "Error creating data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockCourseGetter)

			if tt.expectedStatus != http.StatusBadRequest {
				mockService.On("CreateCourse", mock.Anything, tt.inputCourse).Return(tt.returnedCourse, tt.mockError)
			}
			logger := httplog.NewLogger("test")
			handler := handlers.HandleCreateCourse(logger, mockService)

			req, _ := http.NewRequest("POST", "/api/course", strings.NewReader(tt.requestBody))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response int
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
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

func TestHandleUpdateCourse(t *testing.T) {
	tests := []struct {
		name           string
		courseID       string
		requestBody    string
		inputCourse    models.Course
		returnedCourse models.Course
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Success",
			courseID:       "1",
			requestBody:    `{"Name": "Updated Course"}`,
			inputCourse:    models.Course{Name: "Updated Course"},        // Input to the service
			returnedCourse: models.Course{ID: 1, Name: "Updated Course"}, // Returned by the service
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   models.Course{ID: 1, Name: "Updated Course"}, // Expect full course object
		},
		{
			name:           "Invalid Course ID",
			courseID:       "abc", // Invalid ID
			requestBody:    `{"Name": "Test Course"}`,
			inputCourse:    models.Course{},
			returnedCourse: models.Course{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   handlers.ResponseErr{Error: "Invalid course ID"},
		},
		// {
		// 	name:           "Invalid Request Body",
		// 	courseID:       "1",
		// 	requestBody:    `{}`,
		// 	inputCourse:    models.Course{},
		// 	returnedCourse: models.Course{},
		// 	mockError:      nil,
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedBody:   handlers.ResponseErr{Error: "Invalid request payload"},
		// },
		{
			name:           "Error Updating Course",
			courseID:       "1",
			requestBody:    `{"Name": "Test Course"}`,
			inputCourse:    models.Course{Name: "Test Course"},
			returnedCourse: models.Course{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handlers.ResponseErr{Error: "Error updating data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockCourseGetter)

			// Only mock UpdateCourse when the request is expected to be valid
			if tt.expectedStatus != http.StatusBadRequest {
				courseIDInt, _ := strconv.Atoi(tt.courseID)
				mockService.On("UpdateCourse", mock.Anything, courseIDInt, tt.inputCourse).Return(tt.returnedCourse, tt.mockError)
			}

			logger := httplog.NewLogger("test")
			handler := handlers.HandleUpdateCourse(logger, mockService)

			req, _ := http.NewRequest("PUT", "/api/course/"+tt.courseID, strings.NewReader(tt.requestBody))
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Put("/api/course/{id}", handler)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.Course
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response) // Should match the updated course object
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

func TestHandleDeleteCourse(t *testing.T) {
	tests := []struct {
		name           string
		courseID       string
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Success",
			courseID:       "1",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `"Course has successfully been deleted"`,
		},
		{
			name:           "Invalid Course ID",
			courseID:       "abc", // Invalid ID
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   handlers.ResponseErr{Error: "Invalid course ID"},
		},
		{
			name:           "Error Deleting Course",
			courseID:       "1",
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handlers.ResponseErr{Error: "Error deleting data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockCourseGetter)

			// Only mock DeleteCourse when the request is expected to be valid
			if tt.expectedStatus != http.StatusBadRequest {
				courseIDInt, _ := strconv.Atoi(tt.courseID)
				mockService.On("DeleteCourse", mock.Anything, courseIDInt).Return(tt.mockError)
			}

			logger := httplog.NewLogger("test")
			handler := handlers.HandleDeleteCourse(logger, mockService)

			req, _ := http.NewRequest("DELETE", "/api/course/"+tt.courseID, nil)
			rr := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Delete("/api/course/{id}", handler)
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
