-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, user_name) 
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetUserByName :one
SELECT * from users where user_name = $1;

-- name: Reset :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT * FROM users;