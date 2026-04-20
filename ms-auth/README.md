# ms-auth

`ms-auth` is the authentication and authorization service for Controlitix.

Current status in spec `0018`: service skeleton only.
- Config loading from environment
- Structured logging with `slog` (JSON)
- PostgreSQL connectivity check on startup
- Health endpoints: `/healthz` and `/readyz`
- Graceful shutdown by `SIGINT`/`SIGTERM`

No domain logic, migrations, JWT, repositories, or auth endpoints are implemented in this spec.

## Run Locally

1. Copy `.env.example` to `.env` and set `MS_AUTH_DATABASE_URL`.
2. Start service:

```bash
go run ./cmd/ms-auth
```

Service listens on `:8083` by default.

## Environment

See full list in `.env.example`.

Used in this spec:
- `MS_AUTH_HTTP_ADDRESS` (default `:8083`)
- `MS_AUTH_DATABASE_URL` (required)
- `MS_AUTH_LOG_LEVEL` (default `info`)
- `MS_AUTH_SHUTDOWN_TIMEOUT_SEC` (default `10`)

Reserved variables for future specs are also listed in `.env.example` as comments.

## Health Endpoints

- `GET /healthz` -> `200 {"status":"ok"}`
- `GET /readyz` -> `200 {"status":"ok"}` when PostgreSQL is reachable, otherwise `503 {"status":"degraded","reason":"database unreachable"}`

## Docker

Build image:

```bash
docker build -f Dockerfile .
```

Run with compose:

```bash
docker compose up --build
```

## References

- Architecture: `docs/02_Архитектура/Микросервисы/ms-auth.md`
- ADR-0007: `docs/08_ADR/ADR-0007-service-accounts-env-secrets.md`
- ADR-0012: `docs/08_ADR/ADR-0012-ms-auth-pluggable-identity.md`
