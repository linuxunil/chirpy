
-- name: CreateRefresh :one
insert into refresh_tokens (token, created_at, updated_at, user_id, expires_at)
values ( $1, $2, $3, $4, $5)
returning *;

-- name: GetToken :one
select * from refresh_tokens where token = $1;

-- name: RevokeToken :exec
update refresh_tokens
set revoked_at = $2, updated_at = $2
where token = $1;
