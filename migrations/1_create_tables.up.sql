create extension if not exists "uuid-ossp";

create table accounts (
    id uuid primary key,
    name text not null,
    handle text not null unique,
    password text not null,
    css text,
    avi bytea,
    inserted_at timestamp not null,
    updated_at timestamp not null
);

create index accounts_handle_index on accounts (handle);

create table links (
    id uuid primary key,
    account_id uuid not null,
    title text not null,
    link text not null,
    favicon bytea,
    index integer not null,
    inserted_at timestamp not null,
    updated_at timestamp not null,
    unique (account_id, link),
    foreign key (account_id) references accounts (id) on delete cascade
);
