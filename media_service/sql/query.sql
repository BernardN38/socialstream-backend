-- name: CreateMedia :one
INSERT INTO media(media_id,owner_id)
VALUES ($1,$2) RETURNING *;