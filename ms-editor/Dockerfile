FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /ms-editor ./cmd/ms-editor
RUN go build -o /goose github.com/pressly/goose/v3/cmd/goose

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /ms-editor /app/ms-editor
COPY --from=builder /goose /app/goose
COPY migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/ms-editor"]
