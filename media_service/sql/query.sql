-- name: CreateMedia :one
INSERT INTO media(media_id,owner_id)
VALUES ($1,$2) RETURNING *;

-- name: DeleteMedia :exec
DELETE FROM media WHERE media_id = $1;

