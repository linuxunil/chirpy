// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: users.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING id, created_at, updated_at, email, hashed_password, is_chirpy_red
`

type CreateUserParams struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashed_password"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Email,
		arg.HashedPassword,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
		&i.IsChirpyRed,
	)
	return i, err
}

const getUserAndPassByName = `-- name: GetUserAndPassByName :one
select id, created_at, updated_at, email, hashed_password, is_chirpy_red from users where email = $1
`

func (q *Queries) GetUserAndPassByName(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserAndPassByName, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
		&i.IsChirpyRed,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
select id, created_at, updated_at, email, hashed_password, is_chirpy_red from users where id = $1
`

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
		&i.IsChirpyRed,
	)
	return i, err
}

const getUserByName = `-- name: GetUserByName :one
select id, created_at, updated_at,email from users where email = $1
`

type GetUserByNameRow struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (q *Queries) GetUserByName(ctx context.Context, email string) (GetUserByNameRow, error) {
	row := q.db.QueryRowContext(ctx, getUserByName, email)
	var i GetUserByNameRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
	)
	return i, err
}

const markRed = `-- name: MarkRed :exec
update users set is_chirpy_red = true where id = $1
`

func (q *Queries) MarkRed(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, markRed, id)
	return err
}

const reset = `-- name: Reset :exec
DELETE FROM users chirps
`

func (q *Queries) Reset(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, reset)
	return err
}

const updateUser = `-- name: UpdateUser :one
UPDATE users 
set email = $2, hashed_password = $3
where id = $1
returning id, created_at, updated_at, email, hashed_password, is_chirpy_red
`

type UpdateUserParams struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"hashed_password"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, updateUser, arg.ID, arg.Email, arg.HashedPassword)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
		&i.IsChirpyRed,
	)
	return i, err
}
