-- +goose Up
create extension if not exists ulid;

create table if not exists users (
    id ulid primary key default gen_ulid(),
    username varchar(64),
    email varchar(256) not null unique,
    display_name varchar(256),
    avatar_url varchar(256),
    created_at timestamp with time zone not null default now(),
    updated_at timestamp with time zone not null default now()
);

-- +goose StatementBegin
create or replace function trg_users_set_updated_at()
returns trigger as $$
begin
    new.updated_at = now();
    return new;
end;
$$ language 'plpgsql';
-- +goose StatementEnd

DROP TRIGGER IF EXISTS users_set_updated_at ON users;
CREATE TRIGGER users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION trg_users_set_updated_at();

INSERT INTO users (username, email, display_name, avatar_url)
VALUES ('admin', 'admin@example.com', 'Administrator', NULL)
ON CONFLICT (email) DO NOTHING;


-- +goose Down
drop trigger if exists users_set_updated_at on users;
drop function if exists trg_users_set_updated_at;
drop table if exists users;