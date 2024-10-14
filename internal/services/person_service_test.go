// services/person_service_test.go

package services_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
	"github.com/dchoi22/Go-API-Tech-Challenge/internal/services"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPersonService(t *testing.T) {
	var mockDB *sql.DB

	personService := services.NewPersonService(mockDB)

	require.NotNil(t, personService)
	require.Equal(t, mockDB, personService.Database)
}

func TestGetPeople(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful Get People", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "type", "age", "courses"}).
			AddRow(1, "John", "Doe", "student", 20, pq.Array([]int64{101, 102})).
			AddRow(2, "Jane", "Doe", "student", 22, pq.Array([]int64{103}))

		mock.ExpectQuery(`SELECT p.id, p.first_name, p.last_name, p.type, p.age, ARRAY_AGG\(pc.course_id\) AS courses FROM person p LEFT JOIN person_course pc ON p.id = pc.person_id WHERE type = \$1 GROUP BY id, first_name, last_name, type, age;`).
			WithArgs("student").
			WillReturnRows(rows)

		people, err := service.GetPeople(ctx, "", "", "", "student")
		assert.NoError(t, err)
		assert.Len(t, people, 2)
		assert.Equal(t, "John", people[0].FirstName)
		assert.Equal(t, "Jane", people[1].FirstName)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failed to Get People - Query Error", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectQuery(`SELECT p.id, p.first_name, p.last_name, p.type, p.age, ARRAY_AGG\(pc.course_id\) AS courses FROM person p LEFT JOIN person_course pc ON p.id = pc.person_id WHERE type = \$1 GROUP BY id, first_name, last_name, type, age;`).
			WithArgs("student").
			WillReturnError(errors.New("Database error"))

		_, err := service.GetPeople(ctx, "", "", "", "student")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get people")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failed to Get People - Scan Error", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "type", "age", "courses"}).
			AddRow(1, "John", "Doe", "student", "invalid_age", pq.Array([]int64{101, 102}))

		mock.ExpectQuery(`SELECT p.id, p.first_name, p.last_name, p.type, p.age, ARRAY_AGG\(pc.course_id\) AS courses FROM person p LEFT JOIN person_course pc ON p.id = pc.person_id WHERE type = \$1 GROUP BY id, first_name, last_name, type, age;`).
			WithArgs("student").
			WillReturnRows(rows)

		_, err := service.GetPeople(ctx, "", "", "", "student")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan people from row")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Successful Get People with All Filters", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "type", "age", "courses"}).
			AddRow(1, "John", "Doe", "student", 20, pq.Array([]int64{101, 102}))

		mock.ExpectQuery(`SELECT p.id, p.first_name, p.last_name, p.type, p.age, ARRAY_AGG\(pc.course_id\) AS courses FROM person p LEFT JOIN person_course pc ON p.id = pc.person_id WHERE type = \$1 AND first_name = \$2 AND last_name = \$3 AND age = \$4 GROUP BY id, first_name, last_name, type, age;`).
			WithArgs("student", "John", "Doe", "20").
			WillReturnRows(rows)

		people, err := service.GetPeople(ctx, "John", "Doe", "20", "student")
		assert.NoError(t, err)
		assert.Len(t, people, 1)
		assert.Equal(t, "John", people[0].FirstName)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetPerson(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful Get Person", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "type", "age", "courses"}).
			AddRow(1, "John", "Doe", "student", 20, pq.Array([]int64{101, 102}))

		mock.ExpectQuery(`SELECT p.id, p.first_name, p.last_name, p.type, p.age, ARRAY_AGG\(pc.course_id\) AS courses FROM person p LEFT JOIN person_course pc ON p.id = pc.person_id WHERE "first_name" = \$1 AND "type" = \$2 GROUP BY id, first_name, last_name, type, age;`).
			WithArgs("John", "student").
			WillReturnRows(rows)

		person, err := service.GetPerson(ctx, "John", "student")
		assert.NoError(t, err)
		assert.Equal(t, "John", person.FirstName)
		assert.Equal(t, "Doe", person.LastName)
		assert.Equal(t, 20, person.Age)
		assert.Equal(t, []int64{101, 102}, person.Courses)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failed to Get Person - No Rows", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectQuery(`SELECT p.id, p.first_name, p.last_name, p.type, p.age, ARRAY_AGG\(pc.course_id\) AS courses FROM person p LEFT JOIN person_course pc ON p.id = pc.person_id WHERE "first_name" = \$1 AND "type" = \$2 GROUP BY id, first_name, last_name, type, age;`).
			WithArgs("NonExistent", "student").
			WillReturnError(sql.ErrNoRows)

		_, err := service.GetPerson(ctx, "NonExistent", "student")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get person")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failed to Get Person - Scan Error", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "type", "age", "courses"}).
			AddRow(1, "John", "Doe", "student", "invalid_age", pq.Array([]int64{101, 102})) // Age should be int, not string

		mock.ExpectQuery(`SELECT p.id, p.first_name, p.last_name, p.type, p.age, ARRAY_AGG\(pc.course_id\) AS courses FROM person p LEFT JOIN person_course pc ON p.id = pc.person_id WHERE "first_name" = \$1 AND "type" = \$2 GROUP BY id, first_name, last_name, type, age;`).
			WithArgs("John", "student").
			WillReturnRows(rows)

		_, err := service.GetPerson(ctx, "John", "student")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan person")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestUpdatePersonCourses(t *testing.T) {
	ctx := context.Background()
	studentID := 1
	newCourses := []int64{101, 102, 103}

	t.Run("Success", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM person_course WHERE person_id = \$1 AND course_id != ALL\(\$2\)`).
			WithArgs(studentID, pq.Array(newCourses)).
			WillReturnResult(sqlmock.NewResult(0, 2)) // Assume 2 courses were deleted

		for _, courseID := range newCourses {
			mock.ExpectExec(`INSERT INTO person_course \(person_id, course_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
				WithArgs(studentID, courseID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		mock.ExpectCommit()

		err := service.UpdatePersonCourses(ctx, studentID, newCourses)
		assert.NoError(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failed to Start Transaction", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectBegin().WillReturnError(errors.New("connection error"))

		err := service.UpdatePersonCourses(ctx, studentID, newCourses)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start transaction")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failed to Remove Old Courses", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM person_course WHERE person_id = \$1 AND course_id != ALL\(\$2\)`).
			WithArgs(studentID, pq.Array(newCourses)).
			WillReturnError(errors.New("deletion error"))

		mock.ExpectRollback()

		err := service.UpdatePersonCourses(ctx, studentID, newCourses)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove old courses")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failed to Add New Courses", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM person_course WHERE person_id = \$1 AND course_id != ALL\(\$2\)`).
			WithArgs(studentID, pq.Array(newCourses)).
			WillReturnResult(sqlmock.NewResult(0, 2))

		mock.ExpectExec(`INSERT INTO person_course \(person_id, course_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
			WithArgs(studentID, newCourses[0]).
			WillReturnError(errors.New("insertion error"))

		mock.ExpectRollback()

		err := service.UpdatePersonCourses(ctx, studentID, newCourses)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add new courses")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failed to Commit Transaction", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM person_course WHERE person_id = \$1 AND course_id != ALL\(\$2\)`).
			WithArgs(studentID, pq.Array(newCourses)).
			WillReturnResult(sqlmock.NewResult(0, 2))

		for _, courseID := range newCourses {
			mock.ExpectExec(`INSERT INTO person_course \(person_id, course_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
				WithArgs(studentID, courseID).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}

		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		err := service.UpdatePersonCourses(ctx, studentID, newCourses)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit transaction")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("No Courses to Update", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		emptyNewCourses := []int64{}

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM person_course WHERE person_id = \$1 AND course_id != ALL\(\$2\)`).
			WithArgs(studentID, pq.Array(emptyNewCourses)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectCommit()

		err := service.UpdatePersonCourses(ctx, studentID, emptyNewCourses)
		assert.NoError(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestUpdatePerson(t *testing.T) {
	ctx := context.Background()
	oldFirstName := "John"
	oldPersonType := "Student"
	updatedPerson := models.Person{
		FirstName: "Johnny",
		LastName:  "Doe",
		Type:      "Graduate",
		Age:       25,
	}
	t.Run("Success", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectExec(`
		UPDATE "person" 
		SET "first_name" = \$1, "last_name" = \$2, "type" = \$3, "age" = \$4 
		WHERE "first_name" = \$5 AND "type" = \$6
		`).
			WithArgs(updatedPerson.FirstName, updatedPerson.LastName, updatedPerson.Type, updatedPerson.Age, oldFirstName, oldPersonType).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectQuery(`
		SELECT id, first_name, last_name, type, age 
		FROM "person" 
		WHERE "first_name" = \$1 AND "type" = \$2
		`).
			WithArgs(updatedPerson.FirstName, updatedPerson.Type).
			WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "type", "age"}).
				AddRow(1, updatedPerson.FirstName, updatedPerson.LastName, updatedPerson.Type, updatedPerson.Age))

		result, err := service.UpdatePerson(ctx, oldFirstName, oldPersonType, updatedPerson)

		assert.NoError(t, err)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, updatedPerson.FirstName, result.FirstName)
		assert.Equal(t, updatedPerson.LastName, result.LastName)
		assert.Equal(t, updatedPerson.Type, result.Type)
		assert.Equal(t, updatedPerson.Age, result.Age)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
	t.Run("Update Failure", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectExec(`UPDATE "person" SET "first_name" = \$1, "last_name" = \$2, "type" = \$3, "age" = \$4 WHERE "first_name" = \$5 AND "type" = \$6`).
			WithArgs(updatedPerson.FirstName, updatedPerson.LastName, updatedPerson.Type, updatedPerson.Age, oldFirstName, oldPersonType).
			WillReturnError(errors.New("update failed"))

		_, err := service.UpdatePerson(ctx, oldFirstName, oldPersonType, updatedPerson)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update person")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
	t.Run("Retrieval Failure", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectExec(`UPDATE "person" SET "first_name" = \$1, "last_name" = \$2, "type" = \$3, "age" = \$4 WHERE "first_name" = \$5 AND "type" = \$6`).
			WithArgs(updatedPerson.FirstName, updatedPerson.LastName, updatedPerson.Type, updatedPerson.Age, oldFirstName, oldPersonType).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectQuery(`SELECT id, first_name, last_name, type, age FROM "person" WHERE "first_name" = \$1 AND "type" = \$2`).
			WithArgs(updatedPerson.FirstName, updatedPerson.Type).
			WillReturnError(errors.New("retrieval failed"))

		_, err := service.UpdatePerson(ctx, oldFirstName, oldPersonType, updatedPerson)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve updated person")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("No Rows Returned", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectExec(`UPDATE "person" SET "first_name" = \$1, "last_name" = \$2, "type" = \$3, "age" = \$4 WHERE "first_name" = \$5 AND "type" = \$6`).
			WithArgs(updatedPerson.FirstName, updatedPerson.LastName, updatedPerson.Type, updatedPerson.Age, oldFirstName, oldPersonType).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectQuery(`SELECT id, first_name, last_name, type, age FROM "person" WHERE "first_name" = \$1 AND "type" = \$2`).
			WithArgs(updatedPerson.FirstName, updatedPerson.Type).
			WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "type", "age"}))

		_, err := service.UpdatePerson(ctx, oldFirstName, oldPersonType, updatedPerson)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "[in services.UpdatePerson] failed to retrieve updated person")
		assert.Contains(t, err.Error(), "sql: no rows in result set")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestCreatePerson(t *testing.T) {
	ctx := context.Background()
	t.Run("Success", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectQuery(`
		(?i)^INSERT INTO "person" 
		\(first_name, last_name, type, age\) 
		VALUES \(\$1, \$2, \$3, \$4\) 
		RETURNING id$
		`).
			WithArgs("John", "Smith", "student", 22).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// mock.ExpectCommit()
		_, err := service.CreatePerson(ctx, models.Person{
			FirstName: "John",
			LastName:  "Smith",
			Type:      "student",
			Age:       22,
		})
		assert.NoError(t, err)

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	t.Run("Failed to Create Person", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectQuery(`
		(?i)^INSERT INTO "person" 
		\(first_name, last_name, type, age\) 
		VALUES \(\$1, \$2, \$3, \$4\) 
		RETURNING id$
		`).
			WithArgs("John", "Smith", "student", 22).
			WillReturnError(fmt.Errorf("[in services.CreatePerson] failed to create person"))

		_, err := service.CreatePerson(ctx, models.Person{
			FirstName: "John",
			LastName:  "Smith",
			Type:      "student",
			Age:       22,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "[in services.CreatePerson] failed to create person")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}

	})
}

func TestDeletePerson(t *testing.T) {
	ctx := context.Background()
	t.Run("Success", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()
		mock.ExpectBegin()

		// Mock the person ID query
		mock.ExpectQuery(`
		SELECT id FROM "person" 
		WHERE "first_name" = \$1 AND "type" = \$2
		`).
			WithArgs("John", "student").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// Mock the deletion from person_course
		mock.ExpectExec(`
		DELETE FROM "person_course" 
		WHERE "person_id" = \$1
		`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Mock the deletion from person
		mock.ExpectExec(`
		DELETE FROM "person" 
		WHERE "id" = \$1
		`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		// Call the method under test
		err := service.DeletePerson(ctx, "John", "student")
		assert.NoError(t, err)

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})
	t.Run("Person not found", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		// Mock the transaction begin
		mock.ExpectBegin()

		// Mock the person ID query to return no rows (person not found)
		mock.ExpectQuery(`
		SELECT id FROM "person" 
		WHERE "first_name" = \$1 AND "type" = \$2
		`).
			WithArgs("NonExistent", "student").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// Mock the transaction rollback since the person is not found
		mock.ExpectRollback()

		// Call the method under test
		err := service.DeletePerson(ctx, "NonExistent", "student")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "[in services.DeletePerson] failed to find person: sql: no rows in result set")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})
	t.Run("Failed to Start Transaction", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		// Mock the transaction start failure
		mock.ExpectBegin().WillReturnError(fmt.Errorf("failed to start transaction"))

		// Call the method under test
		err := service.DeletePerson(ctx, "John", "student")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start transaction")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})
	t.Run("Failed to Delete From Person_Course", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectBegin()

		mock.ExpectQuery(`
		SELECT id FROM "person"
        WHERE "first_name" = \$1 AND "type" = \$2
		`).
			WithArgs("John", "student").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectExec(`
		DELETE FROM "person_course" 
		WHERE "person_id" = \$1
		`).
			WithArgs(1).
			WillReturnError(fmt.Errorf("[in services.DeletePerson] failed to delete from person_course"))

		mock.ExpectRollback()

		err := service.DeletePerson(ctx, "John", "student")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete from person_course")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}

	})

	t.Run("No Rows Affected When Deleting Person Record", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectBegin()

		mock.ExpectQuery(`
		SELECT id FROM "person"
        WHERE "first_name" = \$1 AND "type" = \$2
		`).
			WithArgs("John", "student").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectExec(`
		DELETE FROM "person_course" WHERE "person_id" = \$1
		`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`
		DELETE FROM "person" WHERE "id" = \$1
		`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectRollback()

		err := service.DeletePerson(ctx, "John", "student")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	t.Run("Failed to Get Affected Rows", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectBegin()

		mock.ExpectQuery(`
		SELECT id FROM "person"
        WHERE "first_name" = \$1 AND "type" = \$2
		`).
			WithArgs("John", "student").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectExec(`
		DELETE FROM "person_course" WHERE "person_id" = \$1
		`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`
		DELETE FROM "person" WHERE "id" = \$1
		`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("Failed to get affected rows")))

		mock.ExpectRollback()

		err := service.DeletePerson(ctx, "John", "student")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get affected rows")
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

	t.Run("Failed to Commit the Transaction", func(t *testing.T) {
		service, mock := newMockPersonService(t)
		defer service.Database.Close()

		mock.ExpectBegin()

		mock.ExpectQuery(`
		SELECT id FROM "person"
        WHERE "first_name" = \$1 AND "type" = \$2
		`).
			WithArgs("John", "student").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		mock.ExpectExec(`
		DELETE FROM "person_course" WHERE "person_id" = \$1
		`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`
		DELETE FROM "person" WHERE "id" = \$1
		`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit().WillReturnError(fmt.Errorf("failed to commit transaction"))

		err := service.DeletePerson(ctx, "John", "student")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %v", err)
		}
	})

}

func newMockPersonService(t *testing.T) (services.PersonService, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	service := services.PersonService{Database: db}

	return service, mock
}
