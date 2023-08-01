-- +goose Up
ALTER TABLE users
    ADD CONSTRAINT chk_username_length CHECK (LENGTH(username) >= 5),
    ADD CONSTRAINT chk_password_length CHECK (LENGTH(password) >= 8),
    ADD CONSTRAINT chk_email_length CHECK (LENGTH(email) >= 5);

-- +goose Down
ALTER TABLE users
    DROP CONSTRAINT chk_username_length,
    DROP CONSTRAINT chk_email_length,
    DROP CONSTRAINT chk_password_length;
