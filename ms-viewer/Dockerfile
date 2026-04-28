FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /ms-viewer ./cmd/ms-viewer
RUN go build -o /ms-viewer-migrate ./cmd/ms-viewer-migrate

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /ms-viewer /app/ms-viewer
COPY --from=builder /ms-viewer-migrate /app/ms-viewer-migrate

EXPOSE 8084

ENTRYPOINT ["/app/ms-viewer"]
