-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetChirps :many
select * from chirps order by created_at asc;
-- name: GetChirp :one
select * from chirps where id = $1;
