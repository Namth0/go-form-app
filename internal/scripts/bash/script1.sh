#!/bin/bash
# -*- coding: utf-8 -*-
# Script 1 - Attribution de droits utilisateur (Bash)
# Usage: bash script1.sh <user_id>

set -euo pipefail

# Configuration UTF-8 pour tous les systèmes
export LC_ALL=C.UTF-8
export LANG=C.UTF-8
export LANGUAGE=C.UTF-8

# Fonction de logging avec timestamp
log_info() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - INFO - $1"
}

log_error() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - ERROR - $1" >&2
}

# Validation de l'ID utilisateur
validate_user_id() {
    local user_id="$1"
    
    if [[ ! "$user_id" =~ ^[a-zA-Z0-9]{7,12}$ ]]; then
        log_error "Format d'ID utilisateur invalide: $user_id"
        exit 1
    fi
}

# Attribution des droits
grant_permissions() {
    local user_id="$1"
    local permissions=("read_access" "write_access" "execute_access")
    
    log_info "Début attribution des droits pour l'utilisateur: $user_id"
    
    for permission in "${permissions[@]}"; do
        log_info "Attribution du droit '$permission' à $user_id"
        # Simulation de traitement
        sleep 0.1
    done
    
    log_info "Attribution des droits terminée pour $user_id"
}

# Fonction principale
main() {
    if [[ $# -lt 1 ]]; then
        log_error "Usage: bash script1.sh <user_id>"
        exit 1
    fi
    
    local user_id="$1"
    
    log_info "Script Bash 1 démarré pour l'utilisateur: $user_id"
    
    # Validation
    validate_user_id "$user_id"
    
    # Exécution
    grant_permissions "$user_id"
    
    log_info "Script exécuté avec succès"
    echo "SUCCESS: Droits attribués à l'utilisateur $user_id (Bash)"
}

# Exécution si appelé directement
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 