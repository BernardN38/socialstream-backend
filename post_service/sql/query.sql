-- name: GetAll :many
SELECT * FROM posts;

-- name: GetPostUserAndMediaId :one
SELECT user_id, media_id FROM posts WHERE id = $1;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1 AND user_id = $2;

-- name: GetPostPage :many
SELECT * FROM posts WHERE user_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3;

-- name: CreatePost :exec
INSERT INTO Posts(user_id,username,body,media_id) VALUES ($1,$2,$3,$4);
