# ms-editor - agent guide

## Purpose

- Configure monitoring objects, devices, tags, diagrams, figures and publish changes for editor-ui.
- Also owns `devices` and `tags` schemas for device/tag configuration.
- Publishes `config.changed` events to Kafka on entity mutations.
- Owns three DB schemas: `public`, `devices`, `tags`.

## Stack

- Go 1.24 (see `go.mod`).
- Postgres as primary storage.
- Goose for migrations.
- Kafka (segmentio/kafka-go) for event publishing.
- Structured logging via `slog`.

## Architecture (Clean Architecture)

- `cmd/ms-editor` - entrypoint and composition root.
- `internal/domain` - entities and domain errors.
- `internal/usecase` - application use cases.
- `internal/api/http` - HTTP handlers and routing.
- `internal/infrastructure` - database connections and repositories.

## Naming rules

- Use full names only for business variables.
- Technical parameters can use common abbreviations (`ctx`, `req`, `resp`).
- Abbreviations are allowed for technology terms (`http`, `json`, `api`, `url`, `sql`).

## Database rules

- `ms-editor` owns three schemas: `public`, `devices`, `tags`.
- Cross-schema FKs between these three are allowed.
- FKs to schemas owned by other services (e.g. `history`, `alarms`, `auth`) are forbidden.
- All schema changes go through Goose migrations in `migrations/`.
- Do not modify migration files that have already been applied — add a new migration instead.

## API

- Follow `docs/04_API/ms-editor.md` for baseline endpoints.
- Additional endpoints required by editor-ui should be documented in README before implementation.

## Work rules

- One logical change per commit.
- Keep changes scoped to `ms-editor/` repository only.
