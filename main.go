package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"

	httpserver "go-form-app/cmd/server/http"
)

func main() {
	// Configuration du port
	port := os.Getenv("PORT")
	if port == "" {
		// Tentative d'utiliser le port 8001 par défaut
		port = "8001"
		addr := ":" + port
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			// Si le port 8001 est pris, chercher un port libre dans la plage 8001-8015
			found := false
			for i := 0; i < 15; i++ {
				tryPort := 8001 + rand.Intn(15)
				addr = ":" + strconv.Itoa(tryPort)
				ln, err = net.Listen("tcp", addr)
				if err == nil {
					ln.Close()
					port = strconv.Itoa(tryPort)
					found = true
					break
				}
			}
			if !found {
				log.Fatalf("Aucun port disponible dans la plage 8001-8015")
			}
		} else {
			ln.Close()
		}
	}

	// Création et démarrage du serveur sécurisé
	server := httpserver.NewServer()

	log.Printf("Starting Go Form App on port %s", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
