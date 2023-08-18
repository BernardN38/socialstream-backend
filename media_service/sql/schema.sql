CREATE TABLE media
(
    id         serial PRIMARY KEY,
    media_id uuid NOT NULL,
    owner_id int NOT NULL
);
