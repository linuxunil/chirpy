-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;
-- name: GetUserByID :one
select * from users where id = $1;
-- name: GetUserAndPassByName :one
select * from users where email = $1;
-- name: GetUserByName :one
select id, created_at, updated_at,email from users where email = $1;
-- name: Reset :exec
DELETE FROM users chirps;
-- name: UpdateUser :one
UPDATE users 
set email = $2, hashed_password = $3
where id = $1
returning *;
-- name: MarkRed :exec
update users set is_chirpy_red = true where id = $1;
