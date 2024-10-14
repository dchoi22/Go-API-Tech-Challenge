package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
	"github.com/dchoi22/Go-API-Tech-Challenge/internal/services"
	"github.com/stretchr/testify/require"
)

func TestNewCourseService(t *testing.T) {
	var mockDB *sql.DB

	courseService := services.NewCourseService(mockDB)

	require.NotNil(t, courseService)
	require.Equal(t, mockDB, courseService.Database)
}

func TestGetCourses(t *testing.T) {
	service, mock := newMockCourseService(t)
	defer service.Database.Close()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Course 1").
			AddRow(2, "Course 2")
		mock.ExpectQuery(`SELECT \* FROM "course"`).WillReturnRows(rows)

		courses, err := service.GetCourses(context.Background())
		require.NoError(t, err)
		require.Len(t, courses, 2)
		require.Equal(t, courses[0].Name, "Course 1")
		require.Equal(t, courses[1].Name, "Course 2")
	})

	t.Run("QueryError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "course"`).WillReturnError(errors.New("query error"))

		_, err := service.GetCourses(context.Background())
		require.Error(t, err)
	})
}

func TestGetCourse(t *testing.T) {
	service, mock := newMockCourseService(t)
	defer service.Database.Close()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Course 1")
		mock.ExpectQuery(`SELECT \* FROM "course" WHERE "id" = \$1`).WithArgs(1).WillReturnRows(rows)

		course, err := service.GetCourse(context.Background(), 1)
		require.NoError(t, err)
		require.Equal(t, course.Name, "Course 1")
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "course" WHERE "id" = \$1`).WithArgs(1).WillReturnError(sql.ErrNoRows)

		_, err := service.GetCourse(context.Background(), 1)
		require.Error(t, err)
	})

	t.Run("QueryError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "course" WHERE "id" = \$1`).WithArgs(1).WillReturnError(errors.New("query error"))

		_, err := service.GetCourse(context.Background(), 1)
		require.Error(t, err)
	})
}

func TestCreateCourse(t *testing.T) {
	service, mock := newMockCourseService(t)
	defer service.Database.Close()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO "course" \(name\) VALUES \(\$1\) RETURNING "id"`).
			WithArgs("Course 1").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		course, err := service.CreateCourse(context.Background(), models.Course{Name: "Course 1"})
		require.NoError(t, err)
		require.Equal(t, course.ID, 1)
	})

	t.Run("InsertError", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO "course" \(name\) VALUES \(\$1\) RETURNING "id"`).
			WithArgs("Course 1").
			WillReturnError(errors.New("insert error"))

		_, err := service.CreateCourse(context.Background(), models.Course{Name: "Course 1"})
		require.Error(t, err)
	})
}

func TestUpdateCourse(t *testing.T) {
	service, mock := newMockCourseService(t)
	defer service.Database.Close()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "course" WHERE "id" = \$1\)`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		mock.ExpectExec(`UPDATE "course" SET "name" = \$1 WHERE "id" = \$2`).
			WithArgs("Updated Course", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		course, err := service.UpdateCourse(context.Background(), 1, models.Course{Name: "Updated Course"})
		require.NoError(t, err)
		require.Equal(t, course.ID, 1)
	})

	t.Run("CourseNotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "course" WHERE "id" = \$1\)`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		_, err := service.UpdateCourse(context.Background(), 1, models.Course{Name: "Updated Course"})
		require.Error(t, err)
	})

	t.Run("UpdateError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "course" WHERE "id" = \$1\)`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		mock.ExpectExec(`UPDATE "course" SET "name" = \$1 WHERE "id" = \$2`).
			WithArgs("Updated Course", 1).
			WillReturnError(errors.New("update error"))

		_, err := service.UpdateCourse(context.Background(), 1, models.Course{Name: "Updated Course"})
		require.Error(t, err)
	})
}

func TestDeleteCourse(t *testing.T) {
	service, mock := newMockCourseService(t)
	defer service.Database.Close()
	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "course" WHERE "id" = \$1\)`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		mock.ExpectExec(`DELETE FROM "course" WHERE "id" = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := service.DeleteCourse(context.Background(), 1)
		require.NoError(t, err)
	})

	t.Run("CourseNotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "course" WHERE "id" = \$1\)`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		err := service.DeleteCourse(context.Background(), 1)
		require.Error(t, err)
	})

	t.Run("DeleteError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM "course" WHERE "id" = \$1\)`).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		mock.ExpectExec(`DELETE FROM "course" WHERE "id" = \$1`).
			WithArgs(1).
			WillReturnError(errors.New("delete error"))

		err := service.DeleteCourse(context.Background(), 1)
		require.Error(t, err)
	})
}

func newMockCourseService(t *testing.T) (services.CourseService, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	service := services.CourseService{Database: db}

	return service, mock
}
