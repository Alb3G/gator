package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, user_name) 
VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at, user_name
`

type CreateUserParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserName  string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserName,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserName,
	)
	return i, err
}

const getUserByName = `-- name: GetUserByName :one
SELECT user_name from users where user_name = $1
`

func (q *Queries) GetUserByName(ctx context.Context, userName string) (string, error) {
	row := q.db.QueryRowContext(ctx, getUserByName, userName)
	var user_name string
	err := row.Scan(&user_name)
	return user_name, err
}
