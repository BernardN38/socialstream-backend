-- name: GetUserById :one
SELECT *
FROM users
WHERE user_id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserProfileImageByUserId :one
SELECT profile_image_id
FROM users
WHERE user_id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY user_id;

-- name: CreateUser :one
INSERT INTO users(user_id, username,email, firstname,lastname)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateUserProfileImage :exec
UPDATE users SET profile_image_id = $2 WHERE user_id = $1;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE user_id = $1;