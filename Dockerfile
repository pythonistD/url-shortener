FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/server ./cmd/server

FROM alpine:3.21

WORKDIR /app

RUN adduser -D -g '' appuser

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/config.yml /app/config.yml

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/server"]

