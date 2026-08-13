# ============================================
# Stage 1: Build de l'application Vue.js
# ============================================
FROM node:20-alpine AS builder

WORKDIR /app

# Arguments passés au build Vite
ARG VITE_API_BASE_URL
ARG VITE_SERVER_BASE_URL

ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_SERVER_BASE_URL=${VITE_SERVER_BASE_URL}


# Copier les fichiers de dépendances
COPY package*.json ./

# Installer les dépendances
RUN npm ci

# Copier le code source
COPY . .

# Build de production
RUN npm run build

# ============================================
# Stage 2: Serveur Nginx
# ============================================
FROM nginx:alpine

# Supprimer la config par défaut
RUN rm -f /etc/nginx/conf.d/default.conf
RUN rm -f /etc/nginx/nginx.conf

# Copier les fichiers buildés depuis le stage précédent
COPY --from=builder /app/dist /usr/share/nginx/html

# Créer le dossier pour les certificats SSL (montés en volume)
RUN mkdir -p /etc/nginx/ssl

# Exposer les ports HTTP et HTTPS
EXPOSE 80 443

# Commande de démarrage
CMD ["nginx", "-g", "daemon off;"]
