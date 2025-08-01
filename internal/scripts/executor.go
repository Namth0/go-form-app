package scripts

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ScriptType représente le type d'un script
type ScriptType string

const (
	ScriptTypePython ScriptType = "python"
	ScriptTypeBash   ScriptType = "bash"
	ScriptTypeZsh    ScriptType = "zsh"
)

// ExecutionRequest représente une demande d'exécution de script
type ExecutionRequest struct {
	UserID    string
	Script    string
	Arguments []string
}

// ExecutionResult représente le résultat d'une exécution
type ExecutionResult struct {
	Success    bool
	Output     string
	Error      string
	ExitCode   int
	Duration   time.Duration
	ExecutedAt time.Time
}

// Executor gère l'exécution sécurisée des scripts Python, Bash et Zsh
type Executor struct {
	scriptsDir       string
	maxExecutionTime time.Duration
	allowedScripts   []string
	logger           *log.Logger
	userIDPattern    *regexp.Regexp
}

// NewExecutor crée une nouvelle instance de l'executor sécurisé
func NewExecutor(scriptsDir string, maxExecutionTime time.Duration, allowedScripts []string, logger *log.Logger) *Executor {
	return &Executor{
		scriptsDir:       scriptsDir,
		maxExecutionTime: maxExecutionTime,
		allowedScripts:   allowedScripts,
		logger:           logger,
		userIDPattern:    regexp.MustCompile(`^[a-zA-Z0-9]{7,12}$`),
	}
}

// Execute exécute un script Python de manière sécurisée
func (e *Executor) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
	startTime := time.Now()

	if err := e.validateRequest(req); err != nil {
		e.logger.Printf("SECURITY: Request validation failed: %v", err)
		return &ExecutionResult{
			Success:    false,
			Error:      "Invalid request: " + err.Error(),
			ExecutedAt: startTime,
			Duration:   time.Since(startTime),
		}, err
	}

	scriptType := e.detectScriptType(req.Script)
	interpreter := e.getInterpreterCommand(scriptType)
	scriptPath := e.getScriptPath(req.Script, scriptType)

	if !e.isScriptPathSafe(scriptPath) {
		err := fmt.Errorf("script path is not safe: %s", scriptPath)
		e.logger.Printf("SECURITY: %v", err)
		return &ExecutionResult{
			Success:    false,
			Error:      "Script path validation failed",
			ExecutedAt: startTime,
			Duration:   time.Since(startTime),
		}, err
	}

	execCtx, cancel := context.WithTimeout(ctx, e.maxExecutionTime)
	defer cancel()

	args := e.prepareScriptArgs(scriptType, scriptPath, req.UserID)
	args = append(args, req.Arguments...)

	e.logger.Printf("EXECUTION: Starting %s script %s for user %s", scriptType, req.Script, req.UserID)

	cmd := exec.CommandContext(execCtx, interpreter, args...)
	cmd.Env = e.buildSecureEnvironment()

	output, err := cmd.CombinedOutput()

	duration := time.Since(startTime)
	exitCode := cmd.ProcessState.ExitCode()

	result := &ExecutionResult{
		Success:    err == nil && exitCode == 0,
		Output:     e.decodeUTF8Output(output),
		ExitCode:   exitCode,
		Duration:   duration,
		ExecutedAt: startTime,
	}

	if err != nil {
		result.Error = err.Error()
		e.logger.Printf("EXECUTION: Script %s failed for user %s: %v", req.Script, req.UserID, err)
	} else {
		e.logger.Printf("EXECUTION: Script %s completed successfully for user %s (duration: %v)",
			req.Script, req.UserID, duration)
	}

	return result, nil
}

// validateRequest valide la demande d'exécution
func (e *Executor) validateRequest(req ExecutionRequest) error {
	if !e.userIDPattern.MatchString(req.UserID) {
		return fmt.Errorf("invalid user ID format: %s", req.UserID)
	}

	scriptAllowed := false
	for _, allowed := range e.allowedScripts {
		if req.Script == allowed {
			scriptAllowed = true
			break
		}
	}

	if !scriptAllowed {
		return fmt.Errorf("script not in whitelist: %s", req.Script)
	}

	if strings.Contains(req.Script, "..") ||
		strings.Contains(req.Script, "/") ||
		strings.Contains(req.Script, "\\") {
		return fmt.Errorf("invalid characters in script name: %s", req.Script)
	}

	for _, arg := range req.Arguments {
		if e.containsDangerousPatterns(arg) {
			return fmt.Errorf("dangerous pattern detected in argument: %s", arg)
		}
	}

	return nil
}

// isScriptPathSafe vérifie que le chemin du script est sécurisé
func (e *Executor) isScriptPathSafe(scriptPath string) bool {
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return false
	}

	absScriptsDir, err := filepath.Abs(e.scriptsDir)
	if err != nil {
		return false
	}

	if !strings.HasPrefix(absPath, absScriptsDir) {
		return false
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return false
	}

	return true
}

// buildSecureEnvironment construit un environnement d'exécution sécurisé
func (e *Executor) buildSecureEnvironment() []string {
	env := []string{
		"HOME=/tmp",
		"USER=scriptrunner",
		"SHELL=/bin/false",
		"LANG=C.UTF-8",
		"LC_ALL=C.UTF-8",
		"LANGUAGE=C.UTF-8",
		"LC_CTYPE=C.UTF-8",
		"PYTHONIOENCODING=utf-8",
		"PYTHONUNBUFFERED=1",
	}

	isWindows := len(os.Getenv("OS")) > 0 && strings.Contains(strings.ToLower(os.Getenv("OS")), "windows")

	if isWindows {
		env = append(env,
			"PATH=C:\\Windows\\System32;C:\\Windows;C:\\Program Files\\Python39;C:\\Program Files\\Python39\\Scripts",
			"PYTHONLEGACYWINDOWSSTDIO=0",
		)
	} else {
		env = append(env,
			"PATH=/usr/local/bin:/usr/bin:/bin",
		)
	}

	return env
}

// containsDangerousPatterns détecte les patterns dangereux dans les arguments
func (e *Executor) containsDangerousPatterns(arg string) bool {
	dangerousPatterns := []string{
		";", "&", "|", "`", "$", "(", ")", "{", "}", "[", "]",
		"&&", "||", ">>", "<<", "<", ">",
		"rm ", "del ", "format ", "mkfs", "dd ",
		"wget", "curl", "nc ", "netcat",
		"python", "sh", "bash", "cmd", "powershell",
	}

	argLower := strings.ToLower(arg)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(argLower, pattern) {
			return true
		}
	}

	return false
}

// decodeUTF8Output décode correctement la sortie UTF-8 des scripts Python
func (e *Executor) decodeUTF8Output(output []byte) string {
	return string(output)
}

// detectScriptType détecte le type de script basé sur l'extension
func (e *Executor) detectScriptType(scriptName string) ScriptType {
	if strings.HasSuffix(scriptName, ".py") {
		return ScriptTypePython
	} else if strings.HasSuffix(scriptName, ".sh") {
		return ScriptTypeBash
	} else if strings.HasSuffix(scriptName, ".zsh") {
		return ScriptTypeZsh
	}

	return ScriptTypePython
}

// getInterpreterCommand retourne la commande d'interpréteur pour le type de script
func (e *Executor) getInterpreterCommand(scriptType ScriptType) string {
	switch scriptType {
	case ScriptTypeBash:
		return "bash"
	case ScriptTypeZsh:
		return "zsh"
	case ScriptTypePython:
		return "python"
	default:
		return "python"
	}
}

// getScriptPath retourne le chemin complet du script basé sur son type
func (e *Executor) getScriptPath(scriptName string, scriptType ScriptType) string {
	var subDir string

	switch scriptType {
	case ScriptTypeBash:
		subDir = "bash"
	case ScriptTypeZsh:
		subDir = "zsh"
	case ScriptTypePython:
		subDir = "python"
	default:
		subDir = "python"
	}

	return filepath.Join(e.scriptsDir, subDir, scriptName)
}

// prepareScriptArgs prépare les arguments selon le type de script
func (e *Executor) prepareScriptArgs(scriptType ScriptType, scriptPath, userID string) []string {
	switch scriptType {
	case ScriptTypePython:
		return []string{"-u", scriptPath, userID}
	case ScriptTypeBash:
		return []string{scriptPath, userID}
	case ScriptTypeZsh:
		return []string{scriptPath, userID}
	default:
		return []string{scriptPath, userID}
	}
}
