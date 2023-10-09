-- +goose Up
CREATE TABLE media (
    media_id SERIAL PRIMARY KEY,
    external_uuid_full UUID UNIQUE NOT NULL,
    external_uuid_compressed UUID  UNIQUE NOT NULL,
    user_id INT NOT NULL,
    compression_status text NOT NULL,
    upload_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL
);
CREATE INDEX idx_media_minio_uuid ON media(external_uuid_compressed);

-- +goose Down
DROP TABLE media;