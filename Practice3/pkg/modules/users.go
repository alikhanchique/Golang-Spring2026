package modules

import "time"

type User struct {
	ID        int       `db:"id"         json:"id"`
	Name      string    `db:"name"        json:"name"`
	Email     string    `db:"email"       json:"email"`
	Age       int       `db:"age"         json:"age"`
	CreatedAt time.Time `db:"created_at"  json:"created_at"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}
