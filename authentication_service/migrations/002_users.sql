-- +goose Up
ALTER TABLE users
ADD role text;

-- +goose Down
ALTER TABLE users
DROP COLUMN role;