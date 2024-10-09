package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dchoi22/Go-API-Tech-Challenge/internal/models"
	"github.com/lib/pq"
)

type PersonService struct {
	database *sql.DB
}

func NewPersonService(db *sql.DB) *PersonService {
	return &PersonService{
		database: db,
	}
}

func (p PersonService) GetPeople(ctx context.Context, firstName, lastName, age, personType string) ([]models.Person, error) {
	query := `SELECT p.id, p.first_name, p.last_name, p.type, p.age, ARRAY_AGG(pc.course_id) AS courses
		FROM person p
		LEFT JOIN person_course pc ON p.id = pc.person_id
		`
	var whereClauses []string
	var args []interface{}

	whereClauses = append(whereClauses, "type = $"+fmt.Sprint(len(args)+1))
	args = append(args, personType)

	if firstName != "" {
		whereClauses = append(whereClauses, "first_name = $"+fmt.Sprint(len(args)+1))
		args = append(args, firstName)
	}
	if lastName != "" {
		whereClauses = append(whereClauses, "last_name = $"+fmt.Sprint(len(args)+1))
		args = append(args, lastName)
	}
	if age != "" {
		whereClauses = append(whereClauses, "age = $"+fmt.Sprint(len(args)+1))
		args = append(args, age)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += `
	GROUP BY id, first_name, last_name, type, age;
	`

	rows, err := p.database.QueryContext(ctx, query, args...)
	if err != nil {
		return []models.Person{}, fmt.Errorf("[in services.GetPeople] failed to get people: %w", err)
	}
	defer rows.Close()

	var people []models.Person

	for rows.Next() {
		var p models.Person
		err = rows.Scan(&p.ID, &p.FirstName, &p.LastName, &p.Type, &p.Age, pq.Array(&p.Courses))
		if err != nil {
			return []models.Person{}, fmt.Errorf("[in services.GetPeople] failed to scan people from row: %w", err)
		}
		people = append(people, p)
	}
	if err := rows.Err(); err != nil {
		return []models.Person{}, fmt.Errorf("[in services.GetPeople] failed to scan people: %w", err)
	}
	return people, nil
}

func (p PersonService) GetPerson(ctx context.Context, firstName, personType string) (models.Person, error) {
	row := p.database.QueryRowContext(ctx, `
	SELECT p.id, p.first_name, p.last_name, p.type, p.age, ARRAY_AGG(pc.course_id) AS courses
		FROM person p
		LEFT JOIN person_course pc ON p.id = pc.person_id
	WHERE "first_name" = $1 
	AND 
	"type" = $2
	GROUP BY id, first_name, last_name, type, age;
	`, firstName, personType)
	person := models.Person{}
	if err := row.Scan(&person.ID, &person.FirstName, &person.LastName, &person.Type, &person.Age, pq.Array(&person.Courses)); err != nil {
		if err == sql.ErrNoRows {
			return models.Person{}, fmt.Errorf("[in services.GetPerson] failed to get person: %w", err)
		}
		return models.Person{}, fmt.Errorf("[in services.GetCourse] failed to scan person: %w", err)
	}
	return person, nil
}

func (p PersonService) UpdatePerson(ctx context.Context, firstName, personType string, person models.Person) (models.Person, error) {
	_, err := p.database.ExecContext(ctx, `
	UPDATE "person" 
     SET "first_name" = $1, 
         "last_name" = $2, 
         "type" = $3, 
         "age" = $4
     WHERE "first_name" = $5
	 AND "type" = $6;
	 `, person.FirstName, person.LastName, person.Type, person.Age, firstName, personType)
	if err != nil {
		return models.Person{}, fmt.Errorf("[in services.UpdatePerson] failed to update person: %w", err)
	}

	err = p.database.QueryRowContext(ctx, `
        SELECT id, first_name, last_name, type, age FROM "person" 
        WHERE "first_name" = $1 AND "type" = $2;
    `, person.FirstName, person.Type).Scan(
		&person.ID,
		&person.FirstName,
		&person.LastName,
		&person.Type,
		&person.Age,
	)
	if err != nil {
		return models.Person{}, fmt.Errorf("[in services.UpdatePerson] failed to retrieve updated person: %w", err)
	}

	return person, nil
}

func (p PersonService) UpdatePersonCourses(ctx context.Context, studentID int, newCourses []int64) error {

	tx, err := p.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("[in services.UpdatePersonCourses] failed to start transaction: %v", err)
	}

	_, err = tx.ExecContext(ctx, `
        DELETE FROM person_course
        WHERE person_id = $1
        AND course_id != ALL($2)
    `, studentID, pq.Array(newCourses))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("[in services.UpdatePersonCourses] failed to remove old courses: %v", err)
	}

	// Insert the new courses if not already associated with the student
	for _, courseID := range newCourses {
		_, err = tx.ExecContext(ctx, `
            INSERT INTO person_course (person_id, course_id)
            VALUES ($1, $2)
            ON CONFLICT DO NOTHING
        `, studentID, courseID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("[in services.UpdatePersonCourses] failed to add new courses: %v", err)
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("[in services.UpdatePersonCourses] failed to commit transaction: %v", err)
	}

	return nil
}

func (p PersonService) CreatePerson(ctx context.Context, person models.Person) (models.Person, error) {
	err := p.database.QueryRowContext(ctx, `
	INSERT INTO "person" 
	(first_name, last_name, type, age)
	VALUES 
	($1, $2, $3, $4)
	RETURNING id
	`, person.FirstName, person.LastName, person.Type, person.Age).Scan(&person.ID)

	if err != nil {
		return models.Person{}, fmt.Errorf("[in services.CreatePerson] failed to create person: %w", err)
	}
	return person, nil
}

func (p PersonService) DeletePerson(ctx context.Context, firstName, personType string) error {
	result, err := p.database.ExecContext(ctx, `
	DELETE FROM "person"
	WHERE "first_name" = $1
	AND "type" = $2
	`, firstName, personType)
	if err != nil {
		return fmt.Errorf("[in services.DeletePerson] failed to delete person: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("[in services.DeletePerson] failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("person with first name %s and type %s does not exist", firstName, personType)
	}

	return nil
}
