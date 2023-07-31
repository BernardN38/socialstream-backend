-- +goose Up
CREATE TABLE users
(
    id         serial PRIMARY KEY,
    username   varchar(60) NOT NULL UNIQUE,
    email      varchar(60) NOT NULL UNIQUE,
    password   varchar(60) NOT NULL
);

-- +goose Down
DROP TABLE users;