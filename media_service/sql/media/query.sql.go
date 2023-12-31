// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: query.sql

package media_sql

import (
	"context"

	"github.com/google/uuid"
)

const createMedia = `-- name: CreateMedia :one
INSERT INTO media(external_uuid_full,external_uuid_compressed,user_id,compression_status, is_active)
VALUES ($1,$2,$3,$4,$5) RETURNING media_id
`

type CreateMediaParams struct {
	ExternalUuidFull       uuid.UUID `json:"externalUuidFull"`
	ExternalUuidCompressed uuid.UUID `json:"externalUuidCompressed"`
	UserID                 int32     `json:"userId"`
	CompressionStatus      string    `json:"compressionStatus"`
	IsActive               bool      `json:"isActive"`
}

func (q *Queries) CreateMedia(ctx context.Context, arg CreateMediaParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, createMedia,
		arg.ExternalUuidFull,
		arg.ExternalUuidCompressed,
		arg.UserID,
		arg.CompressionStatus,
		arg.IsActive,
	)
	var media_id int32
	err := row.Scan(&media_id)
	return media_id, err
}

const deleteMediaByCompressedExternalId = `-- name: DeleteMediaByCompressedExternalId :exec
DELETE FROM media WHERE external_uuid_compressed = $1
`

func (q *Queries) DeleteMediaByCompressedExternalId(ctx context.Context, externalUuidCompressed uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteMediaByCompressedExternalId, externalUuidCompressed)
	return err
}

const deleteMediaByFullExternalId = `-- name: DeleteMediaByFullExternalId :exec
DELETE FROM media WHERE external_uuid_full = $1
`

func (q *Queries) DeleteMediaByFullExternalId(ctx context.Context, externalUuidFull uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteMediaByFullExternalId, externalUuidFull)
	return err
}

const deleteMediaById = `-- name: DeleteMediaById :exec
DELETE FROM media WHERE media_id = $1
`

func (q *Queries) DeleteMediaById(ctx context.Context, mediaID int32) error {
	_, err := q.db.ExecContext(ctx, deleteMediaById, mediaID)
	return err
}

const deleteMediaByIdWithExternalIds = `-- name: DeleteMediaByIdWithExternalIds :one
DELETE FROM media WHERE media_id = $1 RETURNING external_uuid_full, external_uuid_compressed
`

type DeleteMediaByIdWithExternalIdsRow struct {
	ExternalUuidFull       uuid.UUID `json:"externalUuidFull"`
	ExternalUuidCompressed uuid.UUID `json:"externalUuidCompressed"`
}

func (q *Queries) DeleteMediaByIdWithExternalIds(ctx context.Context, mediaID int32) (DeleteMediaByIdWithExternalIdsRow, error) {
	row := q.db.QueryRowContext(ctx, deleteMediaByIdWithExternalIds, mediaID)
	var i DeleteMediaByIdWithExternalIdsRow
	err := row.Scan(&i.ExternalUuidFull, &i.ExternalUuidCompressed)
	return i, err
}

const getAllMedia = `-- name: GetAllMedia :many
SELECT media_id, external_uuid_full, external_uuid_compressed, user_id, compression_status, upload_date, is_active FROM media
`

func (q *Queries) GetAllMedia(ctx context.Context) ([]Medium, error) {
	rows, err := q.db.QueryContext(ctx, getAllMedia)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Medium
	for rows.Next() {
		var i Medium
		if err := rows.Scan(
			&i.MediaID,
			&i.ExternalUuidFull,
			&i.ExternalUuidCompressed,
			&i.UserID,
			&i.CompressionStatus,
			&i.UploadDate,
			&i.IsActive,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCompressedExternalId = `-- name: GetCompressedExternalId :one
SELECT external_uuid_compressed FROM media WHERE media_id = $1
`

func (q *Queries) GetCompressedExternalId(ctx context.Context, mediaID int32) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getCompressedExternalId, mediaID)
	var external_uuid_compressed uuid.UUID
	err := row.Scan(&external_uuid_compressed)
	return external_uuid_compressed, err
}

const getCompressionStatusById = `-- name: GetCompressionStatusById :one
SELECT compression_status FROM media WHERE media_id = $1
`

func (q *Queries) GetCompressionStatusById(ctx context.Context, mediaID int32) (string, error) {
	row := q.db.QueryRowContext(ctx, getCompressionStatusById, mediaID)
	var compression_status string
	err := row.Scan(&compression_status)
	return compression_status, err
}

const getExternalIdsById = `-- name: GetExternalIdsById :one
SELECT external_uuid_full,external_uuid_compressed, compression_status FROM media WHERE media_id = $1
`

type GetExternalIdsByIdRow struct {
	ExternalUuidFull       uuid.UUID `json:"externalUuidFull"`
	ExternalUuidCompressed uuid.UUID `json:"externalUuidCompressed"`
	CompressionStatus      string    `json:"compressionStatus"`
}

func (q *Queries) GetExternalIdsById(ctx context.Context, mediaID int32) (GetExternalIdsByIdRow, error) {
	row := q.db.QueryRowContext(ctx, getExternalIdsById, mediaID)
	var i GetExternalIdsByIdRow
	err := row.Scan(&i.ExternalUuidFull, &i.ExternalUuidCompressed, &i.CompressionStatus)
	return i, err
}

const getFullExternalId = `-- name: GetFullExternalId :one
SELECT external_uuid_full FROM media WHERE media_id = $1
`

func (q *Queries) GetFullExternalId(ctx context.Context, mediaID int32) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getFullExternalId, mediaID)
	var external_uuid_full uuid.UUID
	err := row.Scan(&external_uuid_full)
	return external_uuid_full, err
}

const updateCompressedExternalId = `-- name: UpdateCompressedExternalId :exec
UPDATE media SET external_uuid_compressed = $2 WHERE external_uuid_compressed = $1
`

type UpdateCompressedExternalIdParams struct {
	ExternalUuidCompressed   uuid.UUID `json:"externalUuidCompressed"`
	ExternalUuidCompressed_2 uuid.UUID `json:"externalUuidCompressed2"`
}

func (q *Queries) UpdateCompressedExternalId(ctx context.Context, arg UpdateCompressedExternalIdParams) error {
	_, err := q.db.ExecContext(ctx, updateCompressedExternalId, arg.ExternalUuidCompressed, arg.ExternalUuidCompressed_2)
	return err
}

const updateCompressedExternalIdByMediaId = `-- name: UpdateCompressedExternalIdByMediaId :exec
UPDATE media
SET external_uuid_compressed = $2
WHERE media_id = $1
`

type UpdateCompressedExternalIdByMediaIdParams struct {
	MediaID                int32     `json:"mediaId"`
	ExternalUuidCompressed uuid.UUID `json:"externalUuidCompressed"`
}

func (q *Queries) UpdateCompressedExternalIdByMediaId(ctx context.Context, arg UpdateCompressedExternalIdByMediaIdParams) error {
	_, err := q.db.ExecContext(ctx, updateCompressedExternalIdByMediaId, arg.MediaID, arg.ExternalUuidCompressed)
	return err
}

const updateCompressionStatus = `-- name: UpdateCompressionStatus :exec
UPDATE media
SET compression_status = $2
WHERE media_id = $1
`

type UpdateCompressionStatusParams struct {
	MediaID           int32  `json:"mediaId"`
	CompressionStatus string `json:"compressionStatus"`
}

func (q *Queries) UpdateCompressionStatus(ctx context.Context, arg UpdateCompressionStatusParams) error {
	_, err := q.db.ExecContext(ctx, updateCompressionStatus, arg.MediaID, arg.CompressionStatus)
	return err
}

const updateExternalIdsById = `-- name: UpdateExternalIdsById :exec
UPDATE media SET  external_uuid_full = $2, external_uuid_compressed = $3 WHERE media_id = $1
`

type UpdateExternalIdsByIdParams struct {
	MediaID                int32     `json:"mediaId"`
	ExternalUuidFull       uuid.UUID `json:"externalUuidFull"`
	ExternalUuidCompressed uuid.UUID `json:"externalUuidCompressed"`
}

func (q *Queries) UpdateExternalIdsById(ctx context.Context, arg UpdateExternalIdsByIdParams) error {
	_, err := q.db.ExecContext(ctx, updateExternalIdsById, arg.MediaID, arg.ExternalUuidFull, arg.ExternalUuidCompressed)
	return err
}

const updateFullExternalId = `-- name: UpdateFullExternalId :exec
UPDATE media SET external_uuid_full = $2 WHERE external_uuid_full = $1
`

type UpdateFullExternalIdParams struct {
	ExternalUuidFull   uuid.UUID `json:"externalUuidFull"`
	ExternalUuidFull_2 uuid.UUID `json:"externalUuidFull2"`
}

func (q *Queries) UpdateFullExternalId(ctx context.Context, arg UpdateFullExternalIdParams) error {
	_, err := q.db.ExecContext(ctx, updateFullExternalId, arg.ExternalUuidFull, arg.ExternalUuidFull_2)
	return err
}

const updateFullExternalIdByMediaId = `-- name: UpdateFullExternalIdByMediaId :exec
UPDATE media
SET external_uuid_full = $2
WHERE media_id = $1
`

type UpdateFullExternalIdByMediaIdParams struct {
	MediaID          int32     `json:"mediaId"`
	ExternalUuidFull uuid.UUID `json:"externalUuidFull"`
}

func (q *Queries) UpdateFullExternalIdByMediaId(ctx context.Context, arg UpdateFullExternalIdByMediaIdParams) error {
	_, err := q.db.ExecContext(ctx, updateFullExternalIdByMediaId, arg.MediaID, arg.ExternalUuidFull)
	return err
}
