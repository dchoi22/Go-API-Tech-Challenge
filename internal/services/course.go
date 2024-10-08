package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
)

type CourseService struct {
	database *sql.DB
}

func NewCourseService(db *sql.DB) *CourseService {
	return &CourseService{
		database: db,
	}
}

func (c CourseService) GetCourses(ctx context.Context) ([]models.Course, error) {
	rows, err := c.database.QueryContext(ctx, `SELECT * FROM "course"`)
	if err != nil {
		return []models.Course{}, fmt.Errorf("[in services.GetCourses] failed to get courses: %w", err)
	}
	defer rows.Close()

	var courses []models.Course

	for rows.Next() {
		var c models.Course
		err = rows.Scan(&c.ID, &c.Name)
		if err != nil {
			return []models.Course{}, fmt.Errorf("[in services.GetCourses] failed to scan courses from row: %w", err)
		}
		courses = append(courses, c)
	}
	if err := rows.Err(); err != nil {
		return []models.Course{}, fmt.Errorf("[in services.GetCourses] failed to scan courses: %w", err)
	}
	return courses, nil
}

func (c CourseService) GetCourse(ctx context.Context, id int) (models.Course, error) {
	row := c.database.QueryRowContext(ctx, `
	SELECT * FROM 
	"course" 
	WHERE "id" = $1
	`, id)
	course := models.Course{}
	if err := row.Scan(&course.ID, &course.Name); err != nil {
		if err == sql.ErrNoRows {
			return models.Course{}, fmt.Errorf("[in services.GetCourse] course not found: %w", err)
		}
		return models.Course{}, fmt.Errorf("[in services.GetCourse] failed to scan course: %w", err)
	}
	return course, nil
}

func (c CourseService) CreateCourse(ctx context.Context, course models.Course) (models.Course, error) {
	err := c.database.QueryRowContext(ctx, `
	INSERT INTO "course" 
	(name) 
	VALUES ($1) 
	RETURNING "id"
	`, course.Name).Scan(&course.ID)
	if err != nil {
		return models.Course{}, fmt.Errorf("[in services.CreateCourse] failed to create course: %w", err)
	}
	return course, nil
}

func (c CourseService) UpdateCourse(ctx context.Context, id int, course models.Course) (models.Course, error) {

	_, err := c.database.ExecContext(ctx, `
	UPDATE "course" 
	SET "name" = $1 
	WHERE "id" = $2
	`, course.Name, id)
	if err != nil {
		return models.Course{}, fmt.Errorf("[in services.UpdateCourse] failed to update course: %w", err)
	}

	course.ID = int(id)
	return course, nil

}
func (c CourseService) DeleteCourse(ctx context.Context, id int) error {
	_, err := c.database.Exec(`
	DELETE FROM 
	"course" 
	WHERE "id" = $1
	`, id)
	if err != nil {
		return fmt.Errorf("[in services.UpdateCourse] failed to update course: %w", err)
	}
	return nil
}
