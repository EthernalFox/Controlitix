module github.com/EthernalFox/Controlitix/ms-viewer

go 1.24.4

require (
	github.com/EthernalFox/Controlitix/shared/authctx v0.0.0
	github.com/coder/websocket v1.8.12
	github.com/go-chi/chi/v5 v5.2.3
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/joho/godotenv v1.5.1
	github.com/pressly/goose/v3 v3.26.0
	github.com/redis/go-redis/v9 v9.7.3
	github.com/segmentio/kafka-go v0.4.47
	golang.org/x/sync v0.16.0
)

replace github.com/EthernalFox/Controlitix/shared/authctx => ../shared/authctx
