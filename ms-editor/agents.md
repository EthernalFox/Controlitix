# ms-editor - agent guide

## Purpose

- Configure monitoring objects, diagrams, figures, and publish changes for editor-ui.
- Owns the `public` schema and its migrations.

## Stack

- Go 1.22 (latest LTS).
- Postgres as primary storage.
- Goose for migrations.
- Structured logging via `slog`.

## Architecture (Clean Architecture)

- `cmd/ms-editor` - entrypoint and composition root.
- `internal/domain` - entities and domain errors.
- `internal/usecase` - application use cases.
- `internal/transport/http` - HTTP handlers and routing.
- `internal/infrastructure` - database connections and repositories.

## Naming rules

- Use full names only for business variables.
- Technical parameters can use common abbreviations (`ctx`, `req`, `resp`).
- Abbreviations are allowed for technology terms (`http`, `json`, `api`, `url`, `sql`).

## Database rules

- Cross-schema foreign keys are forbidden as an architectural rule.
- The service owns tables in `public` and must not create FK constraints to other schemas.
- All schema changes go through Goose migrations in `migrations/`.

## API

- Follow `docs/04_API/ms-editor.md` for baseline endpoints.
- Additional endpoints required by editor-ui should be documented in README before implementation.

## Work rules

- One logical change per commit.
- Keep changes scoped to `ms-editor/` repository only.
