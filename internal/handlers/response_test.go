package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/handlers"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/require"
)

func TestEncodeResponse(t *testing.T) {
	logger := httplog.NewLogger("test", httplog.Options{})

	tests := []struct {
		name           string
		data           any
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Valid JSON Response",
			data: map[string]string{
				"message": "Success",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Success"}`,
		},
		{
			name:           "Internal Server Error for Invalid Data",
			data:           make(chan int), // Invalid data for JSON encoding
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"Error": "Internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handlers.EncodeResponse(rr, logger, tt.expectedStatus, tt.data)

			require.Equal(t, tt.expectedStatus, rr.Code)
			require.JSONEq(t, tt.expectedBody, rr.Body.String())
		})
	}
}
