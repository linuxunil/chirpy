-- +goose up
alter table users 
add column is_chirpy_red boolean not null default FALSE;
