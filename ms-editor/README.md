# ms-editor

Сервис управления объектами мониторинга, мнемосхемами и фигурами для `editor-ui`.

Текущее фактическое покрытие skeleton-проекта:

- объекты мониторинга;
- мнемосхемы;
- фигуры на схеме;
- Swagger и HTTP-каркас.

Устройства, теги и справочники описаны в общей архитектуре, но в этом сервисе ещё не реализованы.

## Стек

- Go 1.24 (актуальная версия)
- Postgres
- Goose migrations
- slog

## Переменные окружения

Параметры подключения вынесены в `.env`. При запуске вне Docker замените хост в
`MS_EDITOR_DATABASE_URL` на `localhost`.

## Локальный запуск

```
go run ./cmd/ms-editor
```

## Миграции

```
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -dir migrations postgres "$MS_EDITOR_DATABASE_URL" up
```

## Docker Compose

```
docker compose up --build
```

## Swagger

Открыть: `http://localhost:8080/swagger`.

## API

Базовые точки: `docs/04_API/ms-editor.md`.

Дополнительно для `editor-ui`:
- `GET /objects`
- `GET /objects/{id}/diagrams`
- `GET /diagrams/{id}/figures`

Поддерживаемые типы фигур синхронизированы с `editor-ui`:

- `rect`
- `circle`
- `ellipse`
- `wedge`
- `line`
- `image`
- `text`
- `ring`
- `arc`
- `tag`
- `path`

## Архитектура

- `internal/domain` - сущности и ошибки
- `internal/usecase` - сценарии
- `internal/api/http` - HTTP слой
- `internal/infrastructure` - БД и репозитории

## Authentication

All editor routes require Bearer access tokens from `ms-auth`.

- `MS_EDITOR_AUTH_JWKS_URL` - JWKS endpoint (for example: `http://ms-auth:8083/.well-known/jwks.json`).
- `MS_EDITOR_AUTH_ISSUER` - expected JWT issuer, default: `controlitix-auth`.
- `MS_EDITOR_AUTH_AUDIENCE` - expected audience, default: `controlitix-api`.
- `MS_EDITOR_AUTH_DISABLED` - dev bypass flag, default: `false`.

Authorization policy:

- all `GET` routes require authenticated token;
- mutating routes (`POST`, `PATCH`, `PUT`, `DELETE`) require role `engineer` or `admin`.

When `MS_EDITOR_AUTH_DISABLED=true`, authentication is bypassed for local development. Do not use this mode in production.
