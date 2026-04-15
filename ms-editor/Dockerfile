FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /ms-editor ./cmd/ms-editor

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /ms-editor /app/ms-editor

EXPOSE 8080

ENTRYPOINT ["/app/ms-editor"]
