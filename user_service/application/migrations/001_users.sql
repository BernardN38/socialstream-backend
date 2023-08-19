-- +goose Up
CREATE TABLE users
(
    user_id         int PRIMARY KEY,
    username   text NOT NULL UNIQUE,
    email      text NOT NULL UNIQUE,
    firstname text NOT NULL,
    lastname text NOT NULL
);
-- +goose Down
DROP TABLE users;