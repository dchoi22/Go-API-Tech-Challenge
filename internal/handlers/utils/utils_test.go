package utils_test

import (
	"testing"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/handlers/utils"
	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
	"github.com/stretchr/testify/require"
)

func TestValidatePerson(t *testing.T) {
	tests := []struct {
		name      string
		person    models.Person
		expectErr string
	}{
		{
			name: "Valid Person",
			person: models.Person{
				FirstName: "John",
				LastName:  "Doe",
				Type:      "student",
				Age:       20,
			},
			expectErr: "",
		},
		{
			name: "Missing First Name",
			person: models.Person{
				FirstName: "",
				LastName:  "Doe",
				Type:      "student",
				Age:       20,
			},
			expectErr: "first name is required",
		},
		{
			name: "Missing Last Name",
			person: models.Person{
				FirstName: "John",
				LastName:  "",
				Type:      "student",
				Age:       20,
			},
			expectErr: "last name is required",
		},
		{
			name: "Invalid Type",
			person: models.Person{
				FirstName: "John",
				LastName:  "Doe",
				Type:      "teacher", // Invalid type
				Age:       20,
			},
			expectErr: "type must be either 'student' or 'professor'",
		},
		{
			name: "Invalid Age",
			person: models.Person{
				FirstName: "John",
				LastName:  "Doe",
				Type:      "student",
				Age:       -1, // Invalid age
			},
			expectErr: "age must be a positive number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidatePerson(tt.person)

			if tt.expectErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tt.expectErr, err.Error())
			}
		})
	}
}
