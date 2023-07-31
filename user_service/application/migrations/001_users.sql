-- +goose Up
CREATE TABLE users
(
    id         serial PRIMARY KEY,
    username   text NOT NULL UNIQUE,
    firstname text NOT NULL,
    lastname text NOT NULL,
    dob date NOT NULL
);

-- +goose Down
DROP TABLE users;