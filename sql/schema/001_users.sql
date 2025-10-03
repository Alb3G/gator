-- +goose Up
CREATE TABLE users(
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	user_name TEXT UNIQUE NOT NULL
);

-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, user_name) 
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetUserByName :one
SELECT user_name from users where user_name = $1;

-- +goose Down
DROP TABLE users;