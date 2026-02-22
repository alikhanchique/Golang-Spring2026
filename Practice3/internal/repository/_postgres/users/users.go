package users

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_postgres "alikhan.practice3/internal/repository/_postgres"
	"alikhan.practice3/pkg/modules"
)

var ErrNotFound = errors.New("user not found")

type Repository struct {
	db               *_postgres.Dialect
	executionTimeout time.Duration
}

func NewUserRepository(db *_postgres.Dialect) *Repository {
	return &Repository{
		db:               db,
		executionTimeout: time.Second * 5,
	}
}

func (r *Repository) GetUsers() ([]modules.User, error) {
	var users []modules.User
	err := r.db.DB.Select(&users, "SELECT id, name, email, age, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("GetUsers: %w", err)
	}
	return users, nil
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	var user modules.User
	err := r.db.DB.Get(&user, "SELECT id, name, email, age, created_at FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetUserByID %d: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("GetUserByID %d: %w", id, err)
	}
	return &user, nil
}

func (r *Repository) CreateUser(req modules.CreateUserRequest) (int, error) {
	if req.Name == "" {
		return 0, errors.New("CreateUser: name is required")
	}
	if req.Email == "" {
		return 0, errors.New("CreateUser: email is required")
	}
	if req.Age <= 0 {
		return 0, errors.New("CreateUser: age must be positive")
	}

	var newID int
	err := r.db.DB.QueryRow(
		`INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id`,
		req.Name, req.Email, req.Age,
	).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: %w", err)
	}
	return newID, nil
}

func (r *Repository) UpdateUser(id int, req modules.UpdateUserRequest) error {
	if req.Name == "" {
		return errors.New("UpdateUser: name is required")
	}
	if req.Email == "" {
		return errors.New("UpdateUser: email is required")
	}
	if req.Age <= 0 {
		return errors.New("UpdateUser: age must be positive")
	}

	result, err := r.db.DB.Exec(
		`UPDATE users SET name = $1, email = $2, age = $3 WHERE id = $4`,
		req.Name, req.Email, req.Age, id,
	)
	if err != nil {
		return fmt.Errorf("UpdateUser %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateUser %d: could not retrieve rows affected: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("UpdateUser %d: %w", id, ErrNotFound)
	}
	return nil
}

func (r *Repository) DeleteUser(id int) (int64, error) {
	result, err := r.db.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return 0, fmt.Errorf("DeleteUser %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("DeleteUser %d: could not retrieve rows affected: %w", id, err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("DeleteUser %d: %w", id, ErrNotFound)
	}
	return rowsAffected, nil
}
