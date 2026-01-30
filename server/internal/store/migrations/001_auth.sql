-- 认证基础表结构

create extension if not exists "uuid-ossp";

create table if not exists users (
  id uuid primary key,
  email varchar(255) not null unique,
  created_at timestamp not null default now()
);

create table if not exists email_verifications (
  id uuid primary key,
  email varchar(255) not null,
  code_hash varchar(255) not null,
  expires_at timestamp not null,
  attempt_count int not null default 0,
  consumed_at timestamp null,
  request_ip text not null,
  created_at timestamp not null default now()
);

create index if not exists idx_email_verifications_email on email_verifications(email);
create index if not exists idx_email_verifications_expires_at on email_verifications(expires_at);

create table if not exists user_identities (
  id uuid primary key,
  user_id uuid not null references users(id),
  provider varchar(50) not null,
  identifier varchar(255) not null,
  created_at timestamp not null default now(),
  unique(provider, identifier)
);
