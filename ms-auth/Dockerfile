FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /ms-auth ./cmd/ms-auth
RUN go build -o /ms-auth-migrate ./cmd/ms-auth-migrate

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /ms-auth /app/ms-auth
COPY --from=builder /ms-auth-migrate /app/ms-auth-migrate

EXPOSE 8083

ENTRYPOINT ["/app/ms-auth"]
