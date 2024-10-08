package utils

import (
	"errors"
	"strings"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
)

func ValidatePerson(person models.Person) error {
	// Validate FirstName
	if strings.TrimSpace(person.FirstName) == "" {
		return errors.New("first name is required")
	}

	// Validate LastName
	if strings.TrimSpace(person.LastName) == "" {
		return errors.New("last name is required")
	}

	// Validate Type (must be either "student" or "professor")
	if person.Type != "student" && person.Type != "professor" {
		return errors.New("type must be either 'student' or 'professor'")
	}

	// Validate Age (must be a positive integer)
	if person.Age <= 0 {
		return errors.New("age must be a positive number")
	}

	// Add other field validations as necessary

	return nil
}
