# ============================================
# Stage 1: Build de l'application Go
# ============================================
FROM golang:1.21-alpine AS builder

# Installer les dépendances de build
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copier les fichiers de dépendances Go
COPY go.mod go.sum ./

# Télécharger les dépendances
RUN go mod download

# Copier le code source
COPY . .

# Compiler l'application pour Linux
# CGO_ENABLED=0 pour un binaire statique
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/api \
    ./cmd/api

# ============================================
# Stage 2: Image de production légère
# ============================================
FROM alpine:3.19

# Installer les certificats CA, timezone, et mariadb-client (fournit les
# binaires "mysql"/"mysqldump" utilisés par la fonctionnalité Sauvegarde,
# internal/backup/service.go).
RUN apk --no-cache add ca-certificates tzdata mariadb-client

# Créer un utilisateur non-root pour la sécurité
RUN adduser -D -g '' appuser

WORKDIR /app

# Copier le binaire depuis le stage de build
COPY --from=builder /build/api .

# Créer les dossiers persistants avec les bonnes permissions (montés en
# volumes par docker-compose)
RUN mkdir -p /app/uploads/covers /app/uploads/avatars /app/backups /app/reports && chown -R appuser:appuser /app

# Utiliser l'utilisateur non-root
USER appuser

# Port exposé
EXPOSE 8080

# Commande de démarrage
CMD ["./api"]
