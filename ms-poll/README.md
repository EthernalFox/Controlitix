# ms-poll

`ms-poll` is a polling service skeleton for Controlitix.

Current scope:
- loads environment config with `MS_POLL_` prefix
- opens PostgreSQL connection (read-only usage)
- initializes Kafka consumer (`config.changed`) and producer (`tags.values`)
- exposes health endpoints (`/healthz`, `/readyz`)
- supports graceful shutdown on `SIGINT`/`SIGTERM`

## Run locally

1. Create `.env` from `.env.example`.
2. Start service:

```bash
go run ./cmd/ms-poll
```

## Environment variables

See full list in `.env.example`.

Required:
- `MS_POLL_DATABASE_URL`
- `MS_POLL_KAFKA_BROKERS`

## Health endpoints

- `GET /healthz` -> liveness, always `200 {"status":"ok"}`
- `GET /readyz` -> readiness, checks DB ping and Kafka brokers config
