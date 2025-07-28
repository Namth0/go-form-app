package utils

import (
	"math/rand"
	"net"
	"os"
	"strconv"
)

// FindAvailablePort trouve un port disponible selon la logique de l'application :
// 1. Utilise la variable d'environnement PORT si définie
// 2. Essaie le port 8001 par défaut
// 3. Si occupé, cherche un port libre dans la plage 8001-8015
func FindAvailablePort() (string, error) {
	// Configuration du port depuis l'environnement
	port := os.Getenv("PORT")
	if port != "" {
		return port, nil
	}

	// Tentative d'utiliser le port 8001 par défaut
	port = "8001"
	if isPortAvailable(port) {
		return port, nil
	}

	// Si le port 8001 est pris, chercher un port libre dans la plage 8001-8015
	for i := 0; i < 15; i++ {
		tryPort := 8001 + rand.Intn(15)
		portStr := strconv.Itoa(tryPort)
		if isPortAvailable(portStr) {
			return portStr, nil
		}
	}

	return "", ErrNoPortAvailable
}

// isPortAvailable vérifie si un port est disponible
func isPortAvailable(port string) bool {
	addr := ":" + port
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// ErrNoPortAvailable est l'erreur retournée quand aucun port n'est disponible
var ErrNoPortAvailable = &PortError{message: "aucun port disponible dans la plage 8001-8015"}

// PortError représente une erreur liée à la gestion des ports
type PortError struct {
	message string
}

func (e *PortError) Error() string {
	return e.message
}
