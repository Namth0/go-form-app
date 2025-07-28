# Multi-stage build pour optimiser la taille de l'image
FROM golang:1.23.4-alpine AS builder

# Installer les dépendances nécessaires
RUN apk add --no-cache git ca-certificates tzdata

# Créer un utilisateur non-root pour la sécurité
RUN adduser -D -s /bin/sh appuser

WORKDIR /app

# Copier les fichiers go.mod et go.sum en premier pour le cache des dépendances
COPY go.mod go.sum ./

# Télécharger les dépendances
RUN go mod download && go mod verify

# Copier le code source
COPY . .

# Exécuter les tests pour s'assurer que tout fonctionne
RUN go test ./...

# Construire l'application avec optimisations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main main.go

# Image finale avec les interpréteurs nécessaires
FROM alpine:3.19

# Installer les interpréteurs et outils nécessaires
RUN apk add --no-cache \
    python3 \
    py3-pip \
    bash \
    zsh \
    wget \
    ca-certificates \
    tzdata \
    && ln -sf python3 /usr/bin/python

# Créer un utilisateur non-root
RUN adduser -D -s /bin/sh appuser

# Copier le binaire
COPY --from=builder /app/main /main

# Copier les fichiers statiques et templates avec la structure attendue
COPY --from=builder /app/cmd/server/http/web /cmd/server/http/web
COPY --from=builder /app/internal/scripts /internal/scripts

# Exposer le port par défaut
EXPOSE 8001

# Utiliser l'utilisateur non-root
USER appuser

# Démarrer l'application
CMD ["/main"]
