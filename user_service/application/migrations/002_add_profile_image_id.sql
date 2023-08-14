-- +goose Up
ALTER TABLE users ADD COLUMN profile_image_id uuid;

-- +goose Down
ALTER TABLE users DROP COLUMN profile_image_id;