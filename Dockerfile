FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/tg_bot ./cmd/bot

FROM alpine:3.21

RUN addgroup -S app && adduser -S -G app app \
    && apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/tg_bot /app/tg_bot
COPY migrations /app/migrations

RUN mkdir -p /app/data /app/tmp && chown -R app:app /app

USER app

ENTRYPOINT ["/app/tg_bot"]
