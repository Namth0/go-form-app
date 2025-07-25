package http

import (
	"log"
	"net/http"
	"os"
	"time"
)

// Server représente le serveur HTTP avec ses configurations
type Server struct {
	handlers *Handlers
	logger   *log.Logger
}

// NewServer crée une nouvelle instance du serveur HTTP
func NewServer() *Server {
	logger := log.New(os.Stdout, "[HTTP-SERVER] ", log.LstdFlags|log.Lshortfile)

	return &Server{
		handlers: NewHandlers(logger),
		logger:   logger,
	}
}

// Start démarre le serveur HTTP avec toutes les protections
func (s *Server) Start(port string) error {
	mux := http.NewServeMux()

	// Routes principales avec middleware de sécurité
	mux.Handle("/", s.securityMiddleware(http.HandlerFunc(s.handlers.FormHandler)))
	mux.Handle("/run-script", s.securityMiddleware(http.HandlerFunc(s.handlers.RunScriptHandler)))

	// Servir les fichiers statiques de manière sécurisée
	staticHandler := http.StripPrefix("/static/",
		http.FileServer(http.Dir("cmd/server/http/web/static/")))
	mux.Handle("/static/", s.securityMiddleware(staticHandler))

	// Configuration du serveur avec timeouts améliorés
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	s.logger.Printf("Starting secure HTTP server on port %s", port)
	return server.ListenAndServe()
}

// securityMiddleware applique les protections de sécurité de base
func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers de sécurité
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' cdn.jsdelivr.net; font-src 'self'; img-src 'self' data: cdn.jsdelivr.net")

		// Rate limiting simple (à améliorer avec Redis en production)
		if !s.checkRateLimit(r) {
			s.logger.Printf("Rate limit exceeded for IP: %s", getClientIP(r))
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		// Log de la requête
		s.logger.Printf("%s %s from %s", r.Method, r.URL.Path, getClientIP(r))

		next.ServeHTTP(w, r)
	})
}

// checkRateLimit vérifie les limites de taux (implémentation basique)
func (s *Server) checkRateLimit(r *http.Request) bool {
	// TODO: Implémenter un vrai rate limiting avec cache/Redis
	// Pour l'instant, on accepte toutes les requêtes
	return true
}
