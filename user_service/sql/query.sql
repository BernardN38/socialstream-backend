-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1 LIMIT 1;


-- name: ListUsers :many
SELECT *
FROM users
ORDER BY id;

-- name: CreateUser :one
INSERT INTO users(username,email, firstname,lastname)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateUserProfileImage :exec
UPDATE users SET profile_image_id = $2 WHERE id = $1;
-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1;