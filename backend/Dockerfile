# ============================================
# Stage 1: Build de l'application Go
# ============================================
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/records-api \
    ./cmd/api

# ============================================
# Stage 2: Image de production légère
# ============================================
FROM alpine:3.19

# mariadb-client fournit les binaires "mysql" et "mysqldump" utilisés par la
# fonctionnalité de sauvegarde/restauration (internal/backup/service.go).
RUN apk --no-cache add ca-certificates tzdata mariadb-client

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /build/records-api .

# Dossiers d'état persistant (montés en volumes par docker-compose)
RUN mkdir -p /app/uploads /app/backups /app/reports \
    && chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./records-api"]
