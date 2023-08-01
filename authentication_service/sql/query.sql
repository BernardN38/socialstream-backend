-- name: GetAllUsers :many
SELECT id, username, email, role
FROM users;

-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserPasswordAndId :one
SELECT id, password
FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserRoleByUserId :one
SELECT role
FROM users
WHERE id = $1 LIMIT 1;


-- name: ListUsers :many
SELECT *
FROM users
ORDER BY id;

-- name: CreateUser :one
INSERT INTO users(username, password, email, role)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1;