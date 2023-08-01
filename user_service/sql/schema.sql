-- +goose Up
CREATE TABLE users
(
    id         serial PRIMARY KEY,
    username   text NOT NULL UNIQUE,
    email      text NOT NULL UNIQUE,
    firstname text NOT NULL,
    lastname text NOT NULL
);