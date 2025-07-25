package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go-form-app/internal/scripts"
)

// SecurityConfig contient les configurations de sécurité
type SecurityConfig struct {
	AllowedScripts   []string
	MaxExecutionTime time.Duration
	UserIDPattern    *regexp.Regexp
	ScriptsDir       string
}

// Handlers contient les handlers HTTP avec les configurations de sécurité
type Handlers struct {
	security SecurityConfig
	logger   *log.Logger
	executor *scripts.Executor
}

// NewHandlers crée une nouvelle instance des handlers avec sécurité
func NewHandlers(logger *log.Logger) *Handlers {
	security := SecurityConfig{
		AllowedScripts: []string{
			// Scripts Python
			"script1.py",
			"script2.py",
			// Scripts Bash
			"script1.sh",
			// Scripts Zsh
			"script1.zsh",
			// Whitelist stricte des scripts autorisés
		},
		MaxExecutionTime: 30 * time.Second,
		UserIDPattern:    regexp.MustCompile(`^[a-zA-Z0-9]{7,12}$`), // Strict pattern pour SSOGF
		ScriptsDir:       "internal/scripts",
	}

	executor := scripts.NewExecutor(
		security.ScriptsDir,
		security.MaxExecutionTime,
		security.AllowedScripts,
		logger,
	)

	return &Handlers{
		security: security,
		logger:   logger,
		executor: executor,
	}
}

// FormHandler affiche le formulaire avec protection CSRF
func (h *Handlers) FormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logSecurityEvent(r, "invalid_method", "GET expected")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Générer un token CSRF sécurisé
	csrfToken, err := generateSecureCSRFToken()
	if err != nil {
		h.logger.Printf("CSRF token generation failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		CSRFToken      string
		AllowedScripts []string
	}{
		CSRFToken:      csrfToken,
		AllowedScripts: h.security.AllowedScripts,
	}

	// Exécution sécurisée du template
	h.executeTemplate(w, "cmd/server/http/web/templates/form.html", data)
}

// RunScriptHandler traite l'exécution des scripts avec validation stricte
func (h *Handlers) RunScriptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logSecurityEvent(r, "invalid_method", "POST expected")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form avec limite de taille
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB max

	// Parser le formulaire multipart si nécessaire
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(1048576); err != nil {
			h.logSecurityEvent(r, "multipart_parse_error", err.Error())
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			h.logSecurityEvent(r, "form_parse_error", err.Error())
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}

	// Validation CSRF - vérifier dans les headers ET dans le form
	csrfToken := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
	if csrfToken == "" {
		csrfToken = strings.TrimSpace(r.FormValue("csrf_token"))
	}

	if csrfToken == "" {
		h.logSecurityEvent(r, "missing_csrf_token", "no token in headers or form")
		h.sendJSONError(w, "Token CSRF manquant", http.StatusBadRequest)
		return
	}

	// TODO: Validation temporelle du token CSRF (pour une sécurité optimale)
	// Pour l'instant, on accepte tout token non-vide

	// Validation stricte des inputs
	userID := strings.TrimSpace(r.FormValue("userId"))
	script := strings.TrimSpace(r.FormValue("script"))

	// Debug: afficher toutes les données reçues
	h.logger.Printf("DEBUG: All form values received:")
	for key, values := range r.Form {
		h.logger.Printf("DEBUG: %s = %v", key, values)
	}
	h.logger.Printf("DEBUG: Raw userID from form: '%s'", r.FormValue("userId"))
	h.logger.Printf("DEBUG: Raw script from form: '%s'", r.FormValue("script"))

	if !h.validateUserID(userID) {
		h.logSecurityEvent(r, "invalid_user_id", userID)
		h.sendJSONError(w, "Format d'ID utilisateur invalide", http.StatusBadRequest)
		return
	}

	if !h.validateScript(script) {
		h.logSecurityEvent(r, "invalid_script", script)
		h.sendJSONError(w, "Script non autorisé", http.StatusBadRequest)
		return
	}

	// Log de l'action avant exécution
	h.logSecurityEvent(r, "script_execution_request",
		fmt.Sprintf("user:%s script:%s", userID, script))

	// Exécution sécurisée du script
	ctx := context.Background()
	req := scripts.ExecutionRequest{
		UserID: userID,
		Script: script,
	}

	result, err := h.executor.Execute(ctx, req)
	if err != nil {
		h.logger.Printf("Script execution failed: %v", err)
		h.sendJSONError(w, "Erreur lors de l'exécution du script", http.StatusInternalServerError)
		return
	}

	// Retourner le résultat
	response := map[string]interface{}{
		"status":   "success",
		"message":  "Script exécuté avec succès",
		"success":  result.Success,
		"duration": result.Duration.String(),
	}

	if !result.Success {
		response["status"] = "error"
		response["message"] = "Échec de l'exécution du script"
		if result.Error != "" {
			response["error"] = result.Error
		}
		if result.Output != "" {
			response["output"] = result.Output
		}
	} else {
		if result.Output != "" {
			response["output"] = result.Output
		}
	}

	// Log du résultat
	h.logSecurityEvent(r, "script_execution_completed",
		fmt.Sprintf("user:%s script:%s success:%t duration:%v exit_code:%d",
			userID, script, result.Success, result.Duration, result.ExitCode))

	h.sendJSONResponse(w, response)
}

// validateUserID valide le format de l'ID utilisateur
func (h *Handlers) validateUserID(userID string) bool {
	if userID == "" {
		h.logger.Printf("DEBUG: Empty userID")
		return false
	}

	// Debug: afficher ce qu'on valide
	h.logger.Printf("DEBUG: Validating userID: '%s' (length: %d)", userID, len(userID))

	isValid := h.security.UserIDPattern.MatchString(userID)
	h.logger.Printf("DEBUG: Pattern match result: %t (pattern: %s)", isValid, h.security.UserIDPattern.String())

	return isValid
}

// validateScript vérifie que le script est dans la whitelist
func (h *Handlers) validateScript(script string) bool {
	if script == "" {
		return false
	}

	// Vérifier contre la whitelist
	for _, allowed := range h.security.AllowedScripts {
		if script == allowed {
			// Vérifier aussi que le nom ne contient pas de path traversal
			if strings.Contains(script, "..") || strings.Contains(script, "/") || strings.Contains(script, "\\") {
				return false
			}
			return true
		}
	}
	return false
}

// sendJSONResponse envoie une réponse JSON
func (h *Handlers) sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Printf("JSON encoding error: %v", err)
	}
}

// sendJSONError envoie une erreur en JSON
func (h *Handlers) sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{
		"status":  "error",
		"message": message,
	}
	json.NewEncoder(w).Encode(response)
}

// logSecurityEvent enregistre les événements de sécurité
func (h *Handlers) logSecurityEvent(r *http.Request, eventType, details string) {
	h.logger.Printf("SECURITY_EVENT: %s | IP: %s | UserAgent: %s | Details: %s",
		eventType,
		getClientIP(r),
		r.UserAgent(),
		details,
	)
}

// getClientIP récupère l'IP réelle du client
func getClientIP(r *http.Request) string {
	// Vérifier les headers de proxy
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}

// generateSecureCSRFToken génère un token CSRF sécurisé
func generateSecureCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// executeTemplate exécute un template de manière sécurisée
func (h *Handlers) executeTemplate(w http.ResponseWriter, templatePath string, data interface{}) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		h.logger.Printf("Template parsing error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		h.logger.Printf("Template execution error: %v", err)
		// Ne pas appeler http.Error si les headers sont déjà envoyés
		// La connexion est probablement fermée côté client
		return
	}
}
