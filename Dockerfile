# ---- Build Stage ----
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# We must generate the gqlgen files inside the container before building
RUN go run github.com/99designs/gqlgen generate

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/bin/api ./cmd/api

# ---- Runtime Stage ----
FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/api /api
COPY --from=builder /app/migrations /migrations

USER 65534:65534
EXPOSE 8080
ENTRYPOINT ["/api"]
