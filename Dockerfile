FROM golang:1.24.3-alpine AS builder

WORKDIR /build

RUN apk update --no-cache && apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

ARG GOOSE_VERSION=v3.24.3
RUN wget -qO goose https://github.com/pressly/goose/releases/download/${GOOSE_VERSION}/goose_linux_x86_64 && \
    chmod +x goose

COPY .env .env
COPY cmd ./cmd
COPY internal ./internal
COPY assets ./assets
COPY scripts ./scripts

RUN GOOS=linux go build -ldflags="-s -w" -o /app ./cmd/bot

FROM alpine:latest

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app /usr/local/bin/app
COPY --from=builder /build/goose /usr/local/bin/goose

COPY --from=builder /build/assets ./assets
COPY --from=builder /build/.env ./.env
COPY --from=builder /build/scripts ./scripts
COPY --from=builder /build/internal/db/migrations ./migrations

ENTRYPOINT ["./scripts/entrypoint.sh"]
