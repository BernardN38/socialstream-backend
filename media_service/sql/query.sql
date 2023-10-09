-- name: CreateMedia :one
INSERT INTO media(external_uuid_full,external_uuid_compressed,user_id,compression_status, is_active)
VALUES ($1,$2,$3,$4,$5) RETURNING media_id;

-- name: GetFullExternalId :one
SELECT external_uuid_full FROM media WHERE media_id = $1;

-- name: GetCompressedExternalId :one
SELECT external_uuid_compressed FROM media WHERE media_id = $1;

-- name: UpdateFullExternalId :exec
UPDATE media SET external_uuid_full = $2 WHERE external_uuid_full = $1;

-- name: GetExternalIdsById :one
SELECT external_uuid_full,external_uuid_compressed, compression_status FROM media WHERE media_id = $1;

-- name: UpdateCompressedExternalId :exec
UPDATE media SET external_uuid_compressed = $2 WHERE external_uuid_compressed = $1;

-- name: UpdateFullExternalIdByMediaId :exec
UPDATE media
SET external_uuid_full = $2
WHERE media_id = $1;

-- name: UpdateExternalIdsById :exec
UPDATE media SET  external_uuid_full = $2, external_uuid_compressed = $3 WHERE media_id = $1;

-- name: UpdateCompressedExternalIdByMediaId :exec
UPDATE media
SET external_uuid_compressed = $2
WHERE media_id = $1;

-- name: DeleteMediaById :exec
DELETE FROM media WHERE media_id = $1;

-- name: DeleteMediaByIdWithExternalIds :one
DELETE FROM media WHERE media_id = $1 RETURNING external_uuid_full, external_uuid_compressed;

-- name: DeleteMediaByFullExternalId :exec
DELETE FROM media WHERE external_uuid_full = $1;

-- name: DeleteMediaByCompressedExternalId :exec
DELETE FROM media WHERE external_uuid_compressed = $1;


-- name: GetCompressionStatusById :one
SELECT compression_status FROM media WHERE media_id = $1;

-- name: UpdateCompressionStatus :exec
UPDATE media
SET compression_status = $2
WHERE media_id = $1;

-- name: GetAllMedia :many
SELECT * FROM media;
