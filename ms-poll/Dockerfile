FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /ms-poll ./cmd/ms-poll

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /ms-poll /app/ms-poll

EXPOSE 8081

ENTRYPOINT ["/app/ms-poll"]
