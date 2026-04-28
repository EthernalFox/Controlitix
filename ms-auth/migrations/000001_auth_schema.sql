-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE auth.identity_sources (
    id serial PRIMARY KEY,
    type text NOT NULL,
    name text NOT NULL,
    config jsonb NOT NULL DEFAULT '{}'::jsonb,
    is_enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT identity_sources_type_check
        CHECK (type IN ('local', 'ldap', 'ad', 'kerberos', 'oidc'))
);

CREATE UNIQUE INDEX identity_sources_name_uq ON auth.identity_sources (name);

CREATE TABLE auth.roles (
    code text PRIMARY KEY,
    name text NOT NULL,
    description text NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE auth.users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    subject text NOT NULL,
    source_id integer NOT NULL REFERENCES auth.identity_sources (id),
    username text NOT NULL,
    display_name text NULL,
    email text NULL,
    password_hash text NULL,
    is_active boolean NOT NULL DEFAULT true,
    last_login_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_subject_uq ON auth.users (subject);
CREATE UNIQUE INDEX users_username_source_uq ON auth.users (source_id, lower(username));
CREATE INDEX users_email_lower_idx ON auth.users (lower(email)) WHERE email IS NOT NULL;

CREATE TABLE auth.user_roles (
    user_id uuid NOT NULL REFERENCES auth.users (id) ON DELETE CASCADE,
    role_code text NOT NULL REFERENCES auth.roles (code),
    granted_at timestamptz NOT NULL DEFAULT now(),
    granted_by uuid NULL REFERENCES auth.users (id),
    PRIMARY KEY (user_id, role_code)
);

CREATE INDEX user_roles_role_idx ON auth.user_roles (role_code);

CREATE TABLE auth.role_mappings (
    id bigserial PRIMARY KEY,
    source_id integer NOT NULL REFERENCES auth.identity_sources (id) ON DELETE CASCADE,
    external_group text NOT NULL,
    role_code text NOT NULL REFERENCES auth.roles (code),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX role_mappings_source_group_role_uq
    ON auth.role_mappings (source_id, lower(external_group), role_code);

CREATE TABLE auth.service_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id text NOT NULL,
    client_secret_hash text NOT NULL,
    display_name text NOT NULL,
    scopes text[] NOT NULL DEFAULT '{}',
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz NULL
);

CREATE UNIQUE INDEX service_accounts_client_id_uq ON auth.service_accounts (client_id);

CREATE TABLE auth.refresh_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES auth.users (id) ON DELETE CASCADE,
    family_id uuid NOT NULL,
    token_hash text NOT NULL,
    parent_id uuid NULL REFERENCES auth.refresh_tokens (id) ON DELETE SET NULL,
    user_agent text NULL,
    ip inet NULL,
    issued_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    used_at timestamptz NULL,
    revoked_at timestamptz NULL,
    revoke_reason text NULL
);

CREATE UNIQUE INDEX refresh_tokens_hash_uq ON auth.refresh_tokens (token_hash);
CREATE INDEX refresh_tokens_user_idx ON auth.refresh_tokens (user_id);
CREATE INDEX refresh_tokens_family_idx ON auth.refresh_tokens (family_id);
CREATE INDEX refresh_tokens_expires_idx ON auth.refresh_tokens (expires_at)
    WHERE revoked_at IS NULL AND used_at IS NULL;

CREATE TABLE auth.audit_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    occurred_at timestamptz NOT NULL DEFAULT now(),
    actor_subject text NULL,
    action text NOT NULL,
    target text NULL,
    result text NOT NULL,
    reason text NULL,
    ip inet NULL,
    user_agent text NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX audit_log_occurred_idx ON auth.audit_log (occurred_at DESC);
CREATE INDEX audit_log_actor_idx ON auth.audit_log (actor_subject);
CREATE INDEX audit_log_action_idx ON auth.audit_log (action);

-- +goose Down
DROP TABLE IF EXISTS auth.audit_log;
DROP TABLE IF EXISTS auth.refresh_tokens;
DROP TABLE IF EXISTS auth.service_accounts;
DROP TABLE IF EXISTS auth.role_mappings;
DROP TABLE IF EXISTS auth.user_roles;
DROP TABLE IF EXISTS auth.users;
DROP TABLE IF EXISTS auth.roles;
DROP TABLE IF EXISTS auth.identity_sources;
DROP SCHEMA IF EXISTS auth;
