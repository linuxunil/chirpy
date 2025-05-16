-- +goose up
ALTER TABLE users ADD COLUMN hashed_password TEXT NOT NULl DEFAULT 'unset';