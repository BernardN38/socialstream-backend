-- +goose Up
ALTER TABLE users
ADD COLUMN role text;

-- +goose Down
ALTER TABLE users
DROP COLUMN role;