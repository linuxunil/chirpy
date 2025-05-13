-- +goose up
create table chirps (
id UUID PRimary key,
	created_at timestamp not null,
	updated_at timestamp not null,
	body text not null,
	user_id UUID NOT NULL references users(id) on delete cascade
);


-- +goose down
drop table chiprs;
