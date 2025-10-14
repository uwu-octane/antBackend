-- +goose Up
CREATE EXTENSION IF NOT EXISTS ulid;

create table if not exists auth_users (
    id ulid primary key default gen_ulid(),
    username varchar(64) unique,
    email varchar(256) not null unique,
    password_hash text not null,
    password_algo varchar(64) not null default 'bcrypt',
    created_at timestamp with time zone not null default now(),
    updated_at timestamp with time zone not null default now()
);

-- +goose StatementBegin
create or replace function trg_auth_users_set_updated_at()
returns trigger as $$
begin
    new.updated_at = now();
    return new;
end;
$$ language 'plpgsql';

Drop Trigger If Exists auth_users_set_updated_at On auth_users;
create Trigger auth_users_set_updated_at
BEFORE UPDATE ON auth_users
FOR EACH ROW
EXECUTE FUNCTION trg_auth_users_set_updated_at();
-- +goose StatementEnd

-- Insert test user: admin/admin123
-- Password hash for 'admin123' using bcrypt
INSERT INTO auth_users (username, email, password_hash, password_algo)
VALUES ('admin', 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'bcrypt')
ON CONFLICT (email) DO NOTHING;

-- +goose Down
DROP TRIGGER IF EXISTS auth_users_set_updated_at ON auth_users;
DROP FUNCTION IF EXISTS trg_auth_users_set_updated_at;
DROP TABLE IF EXISTS auth_users;


