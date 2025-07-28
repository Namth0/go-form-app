# Go Form App - Déploiement Docker

Cette documentation explique comment exécuter la Go Form App en utilisant Docker et Docker Compose.

## 🚀 Démarrage rapide

### Prérequis
- Docker (version 20.x ou plus récente)
- Docker Compose (version 2.x ou plus récente)

### Lancement en mode développement

```bash
# Se placer dans le répertoire docker
cd docker

# Construire et lancer l'application
docker-compose up --build

# Ou en arrière-plan
docker-compose up -d --build
```

L'application sera accessible sur : http://localhost:8001

### Lancement en mode production

```bash
# Lancer avec Nginx en reverse proxy
docker-compose --profile production up -d --build
```

L'application sera accessible sur :
- Port 80 (via Nginx) : http://localhost
- Port 8001 (direct) : http://localhost:8001

## 📋 Commandes utiles

### Gestion des conteneurs

```bash
# Voir les logs
docker-compose logs -f go-form-app

# Arrêter l'application
docker-compose down

# Nettoyer (supprime les volumes)
docker-compose down -v

# Reconstruire sans cache
docker-compose build --no-cache
```

### Debug et maintenance

```bash
# Entrer dans le conteneur (si nécessaire pour debug)
docker-compose exec go-form-app /bin/sh

# Voir le statut des conteneurs
docker-compose ps

# Redémarrer un service
docker-compose restart go-form-app
```

## 🔧 Configuration

### Variables d'environnement

Vous pouvez personnaliser la configuration via le fichier `docker-compose.yml` :

```yaml
environment:
  - PORT=8001          # Port de l'application
  - GO_ENV=development # Environnement (development/production)
```

### Volumes de développement

En mode développement, les volumes sont montés pour permettre le rechargement à chaud :
- `../internal/scripts:/scripts:ro` - Scripts d'exécution
- `../cmd/server/http/web:/web:ro` - Templates et fichiers statiques

## 🛡️ Sécurité

Le Dockerfile implémente plusieurs bonnes pratiques de sécurité :

- **Multi-stage build** : Image finale minimale basée sur `scratch`
- **Utilisateur non-root** : L'application s'exécute sous un utilisateur dédié
- **Image read-only** : Système de fichiers en lecture seule
- **Capabilities limitées** : `no-new-privileges`
- **Timeouts configurés** : Protection contre les attaques DoS

## 🔍 Health Checks

L'application inclut des health checks automatiques :
- Intervalle : 30 secondes
- Timeout : 10 secondes
- Tentatives : 3

## 🌐 Nginx (Production)

En mode production, Nginx agit comme reverse proxy avec :
- Rate limiting (10 req/sec)
- Headers de sécurité
- Compression
- Logs détaillés

## 📁 Structure des fichiers

```
docker/
├── Dockerfile           # Image de l'application
├── docker-compose.yml   # Orchestration des services
├── nginx.conf          # Configuration Nginx
├── .dockerignore       # Fichiers exclus du build
└── README.md           # Cette documentation
```

## 🐛 Résolution de problèmes

### Port déjà utilisé
Si le port 8001 est occupé, l'application trouvera automatiquement un port libre dans la plage 8001-8015.

### Permissions
Vérifiez que Docker a les permissions nécessaires pour accéder aux fichiers du projet.

### Logs
Consultez les logs pour diagnostiquer les problèmes :
```bash
docker-compose logs go-form-app
```

### Rebuild complet
En cas de problème, reconstruisez complètement :
```bash
docker-compose down
docker system prune -a
docker-compose up --build
``` 