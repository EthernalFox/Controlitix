# ms-viewer

Service for trends/history API and ingestion from `tags.values`.

## Build

```bash
go build ./...
```

## Test

```bash
go test ./...
```

## Run

```bash
cp .env.example .env
go run ./cmd/ms-viewer
```

## Migrations

```bash
go run ./cmd/ms-viewer-migrate up
go run ./cmd/ms-viewer-migrate status
go run ./cmd/ms-viewer-migrate version
go run ./cmd/ms-viewer-migrate down
```

## Docker Compose

```bash
docker compose up -d postgres redis kafka
docker compose run --rm ms-viewer-migrate
docker compose up -d ms-viewer
```

Architecture references:
- `../docs/02_Архитектура/Микросервисы/ms-viewer.md`
- `../docs/04_API/ms-viewer.md`
