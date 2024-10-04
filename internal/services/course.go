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

	defer rows.Close()
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

func (c CourseService) GetCourse(ctx context.Context, id string) (models.Course, error) {
	row, err := c.database.QueryContext(ctx, `SELECT * FROM "course" WHERE id =$1`, id)
	if err != nil {
		return models.Course{}, fmt.Errorf("[in services.GetCourse] failed to get course: %w", err)
	}
	course := models.Course{}
	if err = row.Scan(&course.ID, &course.Name); err != nil {
		return models.Course{}, fmt.Errorf("[in services.GetCourse] failed to scan course: %w", err)
	}
	return course, nil
}

func (c CourseService) CreateCourse(ctx context.Context) (models.Course, error) {
	
}