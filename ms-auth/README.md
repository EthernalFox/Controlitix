# ms-auth

`ms-auth` is the authentication and authorization service for Controlitix.

Current status in spec `0018`: service skeleton only.
- Config loading from environment
- Structured logging with `slog` (JSON)
- PostgreSQL connectivity check on startup
- Health endpoints: `/healthz` and `/readyz`
- Graceful shutdown by `SIGINT`/`SIGTERM`

Current status in specs `0018` + `0019`:
- Skeleton HTTP service
- Goose migrations for `auth.*` schema
- Seeds for roles and local identity source
- Optional initial admin seed through environment variables
- `ms-auth-migrate` utility (`up`, `down`, `status`, `hash-password`)

Current status in specs `0018` + `0019` + `0020` + `0021`:
- Domain/auth orchestration (`AuthService`, `IdentityProvider`, `LocalProvider`)
- Argon2id password hasher with pepper
- JWT RS256 keystore and JWKS endpoint at `/.well-known/jwks.json`
- `TokenService` for access/refresh issuing, rotation, and family revocation

Current status in spec `0022`:
- HTTP endpoints: `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`, `GET /api/auth/userinfo`
- RFC 7807 error format (`application/problem+json`)
- In-memory login rate limiter by key `client_ip + username`
- Audit write to PostgreSQL (`auth.audit_log`) with optional Kafka publish to `audit.logs`

Current status in spec `0023`:
- `POST /api/auth/service-token` (client credentials)
- Admin API under `/api/auth/admin/*` with `RequireAuth + RequireRole("admin")`
- Admin user management (CRUD-like with soft deactivate), role assignment, session revoke
- Admin service-account management (create/update/delete, rotate secret)
- Admin audit listing with filters (`action`, `actor`, `from`, `to`)

## Run Locally

1. Copy `.env.example` to `.env` and set `MS_AUTH_DATABASE_URL`.
2. Configure JWT private key via `MS_AUTH_JWT_PRIVATE_KEY_PATH` (preferred) or `MS_AUTH_JWT_PRIVATE_KEY`.
3. Start service:

```bash
go run ./cmd/ms-auth
```

Service listens on `:8083` by default.

## Migrations

Apply migrations:

```bash
MS_AUTH_DATABASE_URL=postgres://... ./ms-auth-migrate up
```

Rollback all migrations:

```bash
MS_AUTH_DATABASE_URL=postgres://... ./ms-auth-migrate down
```

Status:

```bash
MS_AUTH_DATABASE_URL=postgres://... ./ms-auth-migrate status
```

Generate Argon2id hash for initial admin password:

```bash
MS_AUTH_PASSWORD_PEPPER=<base64> echo -n 'secret' | ./ms-auth-migrate hash-password
```

## Environment

See full list in `.env.example`.

Used in this spec:
- `MS_AUTH_HTTP_ADDRESS` (default `:8083`)
- `MS_AUTH_DATABASE_URL` (required)
- `MS_AUTH_LOG_LEVEL` (default `info`)
- `MS_AUTH_SHUTDOWN_TIMEOUT_SEC` (default `10`)
- `MS_AUTH_INITIAL_ADMIN_USERNAME` (optional seed, default `admin`)
- `MS_AUTH_INITIAL_ADMIN_EMAIL` (optional seed)
- `MS_AUTH_INITIAL_ADMIN_PASSWORD_HASH` (optional seed, Argon2id string)
- `MS_AUTH_PASSWORD_PEPPER` (required by `hash-password`)
- `MS_AUTH_JWT_PRIVATE_KEY_PATH` or `MS_AUTH_JWT_PRIVATE_KEY` (required by service startup)
- `MS_AUTH_JWT_KEY_ID` (default `default`)
- `MS_AUTH_JWT_ISSUER` (default `controlitix-auth`)
- `MS_AUTH_JWT_AUDIENCE_USER` (default `controlitix-api`)
- `MS_AUTH_JWT_AUDIENCE_SERVICE` (default `controlitix-internal`)
- `MS_AUTH_JWT_ACCESS_TTL_SEC` (default `900`)
- `MS_AUTH_JWT_REFRESH_TTL_SEC` (default `1209600`)
- `MS_AUTH_LOGIN_RATE_LIMIT` (default `10`)
- `MS_AUTH_LOGIN_RATE_WINDOW_SEC` (default `60`)
- `MS_AUTH_DEV_MODE` (default `false`; when `true`, refresh cookie is sent with `Secure=false`)
- `MS_AUTH_KAFKA_BROKERS` (optional CSV list; empty means audit to DB only)
- `MS_AUTH_KAFKA_AUDIT_TOPIC` (default `audit.logs`)

Reserved variables for future specs are also listed in `.env.example` as comments.

## Health Endpoints

- `GET /healthz` -> `200 {"status":"ok"}`
- `GET /readyz` -> `200 {"status":"ok"}` when PostgreSQL is reachable, otherwise `503 {"status":"degraded","reason":"database unreachable"}`

## JWKS Endpoint

- `GET /.well-known/jwks.json` -> JWKS document with current public signing key (`Cache-Control: public, max-age=900`).

## Auth Endpoints

- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`
- `GET /api/auth/userinfo`
- `POST /api/auth/service-token`
- `GET|POST|PATCH|DELETE /api/auth/admin/*` (see `docs/04_API/ms-auth.md`)

Rate limiter key uses `X-Forwarded-For` first value as client IP (fallback to `RemoteAddr`). This is safe only behind trusted gateway/proxy.

## Docker

Build image:

```bash
docker build -f Dockerfile .
```

Run with compose:

```bash
docker compose up --build
```

Compose runs one-shot `ms-auth-migrate` before starting `ms-auth`.

## References

- Architecture: `docs/02_Архитектура/Микросервисы/ms-auth.md`
- ADR-0007: `docs/08_ADR/ADR-0007-service-accounts-env-secrets.md`
- ADR-0012: `docs/08_ADR/ADR-0012-ms-auth-pluggable-identity.md`
