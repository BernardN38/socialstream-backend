-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserPassword :one
SELECT password
FROM users
WHERE username = $1 LIMIT 1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY id;

-- name: CreateUser :one
INSERT INTO users(username, password, email)
VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1;