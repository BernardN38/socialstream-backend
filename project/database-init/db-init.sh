#!/bin/bash
set -e
export PGPASSWORD=$POSTGRES_PASSWORD
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  CREATE USER $APP_DB_USER WITH PASSWORD '$APP_DB_PASS';
  CREATE DATABASE $AUTH_DB_NAME;
  GRANT ALL PRIVILEGES ON DATABASE $AUTH_DB_NAME TO $APP_DB_USER;
  \connect $AUTH_DB_NAME $APP_DB_USER
  BEGIN;
    CREATE TABLE users
    (
        id         serial PRIMARY KEY,
        username   text NOT NULL UNIQUE,
        email      text NOT NULL UNIQUE,
        password   text NOT NULL
    );
  COMMIT;
EOSQL