FROM golang:1.24.3-alpine AS builder

WORKDIR /build

RUN apk update --no-cache && apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY .env .env
COPY cmd ./cmd
COPY internal ./internal
COPY assets ./assets

RUN GOOS=linux go build -ldflags="-s -w" -o /app ./cmd/bot

FROM alpine:latest

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app /usr/local/bin/app
COPY --from=builder /build/assets ./assets
COPY --from=builder /build/.env ./.env

CMD ["/usr/local/bin/app"]
