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
			"script1.py",
			"script2.py",
			"script1.sh",
			"script1.zsh",
		},
		MaxExecutionTime: 30 * time.Second,
		UserIDPattern:    regexp.MustCompile(`^[a-zA-Z0-9]{7,12}$`),
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

	h.executeTemplate(w, "cmd/server/http/web/templates/form.html", data)
}

// RunScriptHandler traite l'exécution des scripts avec validation stricte
func (h *Handlers) RunScriptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logSecurityEvent(r, "invalid_method", "POST expected")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB max
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

	csrfToken := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
	if csrfToken == "" {
		csrfToken = strings.TrimSpace(r.FormValue("csrf_token"))
	}

	if csrfToken == "" {
		h.logSecurityEvent(r, "missing_csrf_token", "no token in headers or form")
		h.sendJSONError(w, "Token CSRF manquant", http.StatusBadRequest)
		return
	}
	userID := strings.TrimSpace(r.FormValue("userId"))
	script := strings.TrimSpace(r.FormValue("script"))

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

	h.logSecurityEvent(r, "script_execution_request",
		fmt.Sprintf("user:%s script:%s", userID, script))
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

	h.logSecurityEvent(r, "script_execution_completed",
		fmt.Sprintf("user:%s script:%s success:%t duration:%v exit_code:%d",
			userID, script, result.Success, result.Duration, result.ExitCode))

	h.sendJSONResponse(w, response)
}

// validateUserID valide le format de l'ID utilisateur
func (h *Handlers) validateUserID(userID string) bool {
	if userID == "" {
		return false
	}

	return h.security.UserIDPattern.MatchString(userID)
}

// validateScript vérifie que le script est dans la whitelist
func (h *Handlers) validateScript(script string) bool {
	if script == "" {
		return false
	}

	for _, allowed := range h.security.AllowedScripts {
		if script == allowed {
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
		return
	}
}
