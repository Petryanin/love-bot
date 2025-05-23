FROM golang:1.24.3-alpine AS builder

WORKDIR /build

RUN apk update --no-cache && apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

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
COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY --from=builder /build/assets ./assets
COPY --from=builder /build/.env ./.env
COPY --from=builder /build/scripts ./scripts
COPY --from=builder /build/internal/db/migrations ./migrations

ENTRYPOINT ["./scripts/entrypoint.sh"]
