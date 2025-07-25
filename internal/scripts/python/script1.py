#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Script 1 - Exemple de script pour attribution de droits
Usage: python script1.py <user_id>
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

def grant_permissions(user_id):
    """Simule l'attribution de droits pour un utilisateur"""
    logger = logging.getLogger(__name__)
    
    # Simulation d'attribution de droits
    permissions = [
        "read_access",
        "write_access", 
        "execute_access"
    ]
    
    logger.info(f"Début attribution des droits pour l'utilisateur: {user_id}")
    
    for permission in permissions:
        # Simulation de traitement
        logger.info(f"Attribution du droit '{permission}' à {user_id}")
        
    logger.info(f"Attribution des droits terminée pour {user_id}")
    return True

def main():
    """Fonction principale"""
    logger = setup_logging()
    
    # Vérification des arguments
    if len(sys.argv) < 2:
        logger.error("Usage: python script1.py <user_id>")
        sys.exit(1)
    
    user_id = sys.argv[1].strip()
    
    # Validation de l'ID utilisateur
    if not validate_user_id(user_id):
        logger.error(f"Format d'ID utilisateur invalide: {user_id}")
        sys.exit(1)
    
    logger.info(f"Script 1 démarré pour l'utilisateur: {user_id}")
    
    try:
        # Exécution de l'attribution des droits
        success = grant_permissions(user_id)
        
        if success:
            logger.info("Script exécuté avec succès")
            print(f"SUCCESS: Droits attribués à l'utilisateur {user_id}")
            sys.exit(0)
        else:
            logger.error("Échec de l'attribution des droits")
            sys.exit(1)
            
    except Exception as e:
        logger.error(f"Erreur lors de l'exécution: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    main() 