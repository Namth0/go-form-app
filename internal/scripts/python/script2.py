#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Script 2 - Exemple de script pour gestion d'accès avancé
Usage: python script2.py <user_id>
"""
import sys
import re
import logging
from datetime import datetime
import os

# Force UTF-8 encoding for all platforms
import codecs

# Configure UTF-8 for all systems
if hasattr(sys.stdout, 'detach'):
    try:
        sys.stdout = codecs.getwriter('utf-8')(sys.stdout.detach())
        sys.stderr = codecs.getwriter('utf-8')(sys.stderr.detach())
    except:
        # Fallback si detach() n'est pas disponible
        sys.stdout = codecs.getwriter('utf-8')(sys.stdout.buffer)
        sys.stderr = codecs.getwriter('utf-8')(sys.stderr.buffer)

def setup_logging():
    """Configuration du logging sécurisé"""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(levelname)s - %(message)s',
        handlers=[
            logging.StreamHandler(sys.stdout)
        ]
    )
    return logging.getLogger(__name__)

def validate_user_id(user_id):
    """Valide le format de l'ID utilisateur"""
    if not user_id:
        return False
    
    # Pattern strict pour SSOGF
    pattern = re.compile(r'^[a-zA-Z0-9]{7,12}$')
    return pattern.match(user_id) is not None

def configure_advanced_access(user_id):
    """Simule la configuration d'accès avancé pour un utilisateur"""
    logger = logging.getLogger(__name__)
    
    # Simulation de configuration avancée
    configurations = [
        "database_access_level_2",
        "api_access_premium", 
        "admin_panel_access",
        "reporting_access"
    ]
    
    logger.info(f"Début configuration avancée pour l'utilisateur: {user_id}")
    
    for config in configurations:
        # Simulation de traitement plus long
        logger.info(f"Configuration '{config}' appliquée à {user_id}")
        
    logger.info(f"Configuration avancée terminée pour {user_id}")
    return True

def main():
    """Fonction principale"""
    logger = setup_logging()
    
    # Vérification des arguments
    if len(sys.argv) < 2:
        logger.error("Usage: python script2.py <user_id>")
        sys.exit(1)
    
    user_id = sys.argv[1].strip()
    
    # Validation de l'ID utilisateur
    if not validate_user_id(user_id):
        logger.error(f"Format d'ID utilisateur invalide: {user_id}")
        sys.exit(1)
    
    logger.info(f"Script 2 démarré pour l'utilisateur: {user_id}")
    
    try:
        # Exécution de la configuration avancée
        success = configure_advanced_access(user_id)
        
        if success:
            logger.info("Script exécuté avec succès")
            print(f"SUCCESS: Configuration avancée appliquée à l'utilisateur {user_id}")
            sys.exit(0)
        else:
            logger.error("Échec de la configuration")
            sys.exit(1)
            
    except Exception as e:
        logger.error(f"Erreur lors de l'exécution: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    main() 