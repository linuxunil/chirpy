-- +goose Up
create table users (
ID UUID PRIMARY KEY,
created_at Timestamp not null,
	updated_at timestamp not null,
	email text not null
);

-- +goose down
drop table users;
