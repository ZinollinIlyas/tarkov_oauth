CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    id            uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname      VARCHAR(100) UNIQUE,
    display_name  VARCHAR(100) NOT NULL,
    password      text,
    created_at    timestamptz  NOT NULL DEFAULT now(),
    updated_at    timestamptz  NOT NULL DEFAULT now(),
    last_login_at timestamptz,
    disabled_at   timestamptz
);
