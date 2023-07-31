-- +goose Up
alter table users
  add constraint check_min_length_username check (length(username) >= 8);
alter table users
  add constraint check_min_length_email check (length(email) >= 8);
alter table users
  add constraint check_min_length_password check (length(password) >= 8);

-- +goose Down
ALTER TABLE users DROP CONSTRAINT check_min_length_username; 
ALTER TABLE users DROP CONSTRAINT check_min_length_email; 
ALTER TABLE users DROP CONSTRAINT check_min_length_password; 