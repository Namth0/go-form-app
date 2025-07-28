# Go Form App

Une application web sécurisée en Go permettant l'exécution contrôlée de scripts Python, Bash et Zsh via une interface web.

## Fonctionnalités

- Interface web simple et sécurisée
- Exécution de scripts Python, Bash et Zsh
- Validation stricte des entrées utilisateur
- Protection CSRF
- Logs de sécurité détaillés
- Isolation des environnements d'exécution
- Support Docker complet

## Architecture

```
go-form-app/
├── cmd/server/http/          # Serveur HTTP et handlers
├── internal/
│   ├── scripts/             # Moteur d'exécution des scripts
│   └── utils/               # Utilitaires (gestion des ports)
├── docker/                  # Configuration Docker
└── main.go                  # Point d'entrée de l'application
```

## Installation et Démarrage

### Prérequis

- Go 1.21 ou plus récent
- Docker et Docker Compose (pour le déploiement containerisé)

### Développement Local

```bash
# Cloner le projet
git clone <repository-url>
cd go-form-app

# Installer les dépendances
go mod download

# Lancer l'application
go run main.go
```

L'application sera accessible sur http://localhost:8001

### Déploiement Docker

```bash
# Mode développement
cd docker
docker-compose up --build

# Mode production avec Nginx
docker-compose --profile production up -d --build
```

## Configuration

### Variables d'environnement

- `PORT`: Port d'écoute (défaut: 8001, avec recherche automatique de port libre)
- `GO_ENV`: Environnement d'exécution (development/production)

### Scripts autorisés

Les scripts autorisés sont définis dans la whitelist du code :
- `script1.py`, `script2.py` (Python)
- `script1.sh` (Bash)
- `script1.zsh` (Zsh)

## Sécurité

### Mesures implémentées

- **Validation stricte des entrées**: Format UserID, whitelist des scripts
- **Protection CSRF**: Tokens sécurisés pour toutes les requêtes POST
- **Isolation d'exécution**: Environnement limité, timeouts configurés
- **Protection path traversal**: Blocage des tentatives d'accès aux fichiers système
- **Détection d'injection**: Filtrage des patterns dangereux dans les arguments
- **Headers de sécurité**: X-Frame-Options, CSP, X-XSS-Protection
- **Rate limiting**: Protection contre les attaques DoS
- **Utilisateur non-root**: Exécution avec privilèges limités

### Format UserID

Les identifiants utilisateur doivent respecter le format :
- 7 à 12 caractères alphanumériques uniquement
- Pas de caractères spéciaux ou espaces

## Tests

### Exécution des tests

```bash
# Tests unitaires
make test

# Tests avec verbose
make test-verbose

# Rapport de couverture
make test-coverage

# Benchmarks
make bench
```

### Couverture de tests

| Package | Coverage | Status |
|---------|----------|--------|
| `internal/utils` | **89.5%** ✅ | Excellent |
| `internal/scripts` | **67.0%** ✅ | Très bon |
| `cmd/server/http` | **58.8%** ⚠️ | Correct |
| **Global** | **62.5%** ✅ | Bon |

## Développement

### Makefile

Le projet inclut un Makefile avec les commandes courantes :

```bash
make help              # Aide
make build             # Compiler l'application
make test              # Tests unitaires
make test-coverage     # Rapport de couverture HTML
make docker-build      # Build Docker
make docker-run        # Lancer avec Docker
make clean             # Nettoyer les artefacts
```

### Structure du code

- `main.go`: Point d'entrée, gestion des ports
- `cmd/server/http/`: Serveur HTTP, handlers, middleware de sécurité
- `internal/scripts/`: Moteur d'exécution sécurisé des scripts
- `internal/utils/`: Utilitaires réutilisables (gestion des ports)

## API

### Endpoints

- `GET /`: Interface web (formulaire)
- `POST /run-script`: Exécution de script (requiert CSRF token)
- `GET /static/*`: Fichiers statiques

### Format de requête

```bash
POST /run-script
Content-Type: application/x-www-form-urlencoded

userId=abc1234&script=script1.py&csrf_token=<token>
```

### Réponse

```json
{
  "status": "success|error",
  "message": "Description",
  "output": "Sortie du script",
  "duration": "Temps d'exécution"
}
```

## Logs

### Types de logs

- **EXECUTION**: Démarrage et fin d'exécution des scripts
- **SECURITY_EVENT**: Événements de sécurité (tentatives d'intrusion, validations)
- **DEBUG**: Informations de débogage (validation des entrées)

### Exemple de logs

```
[HTTP-SERVER] 2025/01/01 12:00:00 SECURITY_EVENT: script_execution_request | IP: 192.168.1.1 | UserAgent: Mozilla/5.0 | Details: user:abc1234 script:script1.py
[HTTP-SERVER] 2025/01/01 12:00:01 EXECUTION: Script script1.py completed successfully for user abc1234 (duration: 1.2s)
```

## Monitoring

### Health Checks

L'application inclut des health checks pour Docker :
- Endpoint: GET /
- Intervalle: 30 secondes
- Timeout: 10 secondes

### Métriques

- Temps d'exécution des scripts
- Taux de succès/échec
- Événements de sécurité
- Performance des validations

## Contribution

### Standards de code

- Tests obligatoires pour les nouvelles fonctionnalités
- Couverture minimale: 60% par package
- Linting avec golangci-lint
- Documentation des fonctions publiques

### Process

1. Fork le projet
2. Créer une branche feature
3. Ajouter des tests
4. Vérifier la couverture
5. Soumettre une Pull Request

## Licence

[Spécifier la licence du projet]

## Support

Pour le support technique :
1. Vérifier les logs de l'application
2. Consulter la section "Résolution de problèmes" dans docker/README.md
3. Ouvrir une issue avec les logs pertinents