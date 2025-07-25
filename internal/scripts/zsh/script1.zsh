#!/bin/zsh
# -*- coding: utf-8 -*-
# Script 1 - Configuration avancée utilisateur (Zsh)
# Usage: zsh script1.zsh <user_id>

setopt ERR_EXIT
setopt PIPE_FAIL

# Configuration UTF-8 pour tous les systèmes
export LC_ALL=C.UTF-8
export LANG=C.UTF-8
export LANGUAGE=C.UTF-8

# Fonction de logging avec timestamp et couleurs
log_info() {
    print "$(date '+%Y-%m-%d %H:%M:%S') - INFO - $1"
}

log_error() {
    print "$(date '+%Y-%m-%d %H:%M:%S') - ERROR - $1" >&2
}

log_success() {
    print "$(date '+%Y-%m-%d %H:%M:%S') - SUCCESS - $1"
}

# Validation de l'ID utilisateur avec Zsh regex
validate_user_id() {
    local user_id="$1"
    
    if [[ ! "$user_id" =~ '^[a-zA-Z0-9]{7,12}$' ]]; then
        log_error "Format d'ID utilisateur invalide: $user_id"
        exit 1
    fi
}

# Configuration avancée des accès
configure_advanced_access() {
    local user_id="$1"
    local -a configurations=(
        "database_access_level_2"
        "api_access_premium"
        "admin_panel_access"
        "reporting_access"
        "monitoring_access"
    )
    
    log_info "Début configuration avancée pour l'utilisateur: $user_id"
    
    for config in $configurations; do
        log_info "Configuration '$config' appliquée à $user_id"
        # Simulation de traitement plus complexe
        sleep 0.15
    done
    
    log_success "Configuration avancée terminée pour $user_id"
}

# Vérification des prérequis système
check_prerequisites() {
    local user_id="$1"
    
    log_info "Vérification des prérequis système pour $user_id"
    
    # Simulation de vérifications
    local -a checks=(
        "Vérification base de données"
        "Validation API endpoints"
        "Test connectivité réseau"
        "Contrôle permissions système"
    )
    
    for check in $checks; do
        log_info "$check - OK"
        sleep 0.1
    done
}

# Fonction principale
main() {
    if [[ $# -lt 1 ]]; then
        log_error "Usage: zsh script1.zsh <user_id>"
        exit 1
    fi
    
    local user_id="$1"
    
    log_info "Script Zsh 1 démarré pour l'utilisateur: $user_id"
    
    # Validation
    validate_user_id "$user_id"
    
    # Vérifications
    check_prerequisites "$user_id"
    
    # Configuration
    configure_advanced_access "$user_id"
    
    log_success "Script exécuté avec succès"
    print "SUCCESS: Configuration avancée appliquée à l'utilisateur $user_id (Zsh)"
}

# Exécution si appelé directement
if [[ "${ZSH_EVAL_CONTEXT}" == "toplevel" ]]; then
    main "$@"
fi 