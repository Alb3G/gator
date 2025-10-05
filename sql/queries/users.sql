-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, user_name) 
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetUserByName :one
SELECT user_name from users where user_name = $1;