package main

import (
	"log"

	httpserver "go-form-app/cmd/server/http"
	"go-form-app/internal/utils"
)

func main() {
	// Recherche d'un port disponible
	port, err := utils.FindAvailablePort()
	if err != nil {
		log.Fatalf("Erreur lors de la recherche de port: %v", err)
	}

	// Création et démarrage du serveur sécurisé
	server := httpserver.NewServer()

	log.Printf("Starting Go Form App on port %s", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
