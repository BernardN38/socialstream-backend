-- +goose Up
CREATE TABLE posts
(
    id SERIAL PRIMARY KEY,
    user_id int NOT NULL, 
    username text NOT NULL,   
    body text NOT NULL, 
    media_id int,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE posts;