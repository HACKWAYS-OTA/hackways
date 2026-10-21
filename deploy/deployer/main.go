package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Deployment represents a single deployment record.
type Deployment struct {
	ID         string    `json:"id"`
	Commit     string    `json:"commit"`
	ShortSHA   string    `json:"short_sha"`
	Message    string    `json:"message"`
	Author     string    `json:"author"`
	Branch     string    `json:"branch"`
	Generation uint64    `json:"generation"`
	Trigger    string    `json:"trigger"`
	Status     string    `json:"status"` // "BUILDING", "READY", or "FAILED"
	CreatedAt  time.Time `json:"created_at"`
	EndedAt    time.Time `json:"ended_at,omitempty"`
	Duration   string    `json:"duration,omitempty"`
}

// Deployer deploys CI-verified commits and accepts signed GitHub webhooks.
type Deployer struct {
	repoDir       string
	branch        string
	webhookSecret string
	mu            sync.Mutex
	isDeploying   bool
	autoEnabled   bool
	repository    string
	approvalFile  string
	stateFile     string
	active        *Deployment
	pending       *Deployment
	pipeline      func(*Deployment) error
	logger        *slog.Logger
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	repoDir := getEnv("REPO_DIR", "/opt/hackways")
	branch := getEnv("BRANCH", "main")
	secret := getEnv("WEBHOOK_SECRET", "")
	if strings.TrimSpace(secret) == "" {
		logger.Error("WEBHOOK_SECRET is required; refusing to start unsigned deployment webhook")
		os.Exit(1)
	}
	port := 9898

	deployer := &Deployer{
		repoDir:       repoDir,
		branch:        branch,
		webhookSecret: secret,
		logger:        logger,
		autoEnabled:   getEnv("AUTO_DEPLOY_ENABLED", "false") == "true",
		repository:    getEnv("GITHUB_REPOSITORY", "HACKWAYS-OTA/hackways"),
		approvalFile:  getEnv("APPROVED_RELEASE_FILE", "/etc/hackways/approved-release.json"),
		stateFile:     getEnv("DEPLOY_STATE_FILE", "/var/lib/hackways-deployer/state.json"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/deploy/webhook", deployer.handleWebhook)
	mux.HandleFunc("/health", deployer.handleHealth)

	addr := fmt.Sprintf("%s:%d", getEnv("BIND_HOST", "127.0.0.1"), port)
	logger.Info("Hackways CI deployment webhook starting",
		"port", port,
		"repo_dir", repoDir,
		"branch", branch,
		"mode", "signed CI completion webhook",
	)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func shortSHA(commit string) string {
	if len(commit) <= 7 {
		return commit
	}
	return commit[:7]
}

func (d *Deployer) runDeployPipeline(dep *Deployment) {
	startTime := time.Now()

	appendLog := func(format string, a ...interface{}) {
		msg := fmt.Sprintf(format, a...)
		d.logger.Info("deploy log", "dep_id", dep.ID, "line", msg)
	}

	appendLog("Hackways deployment started")
	appendLog("Branch: %s | CI image tag: %s | Trigger: %s", dep.Branch, dep.Commit, dep.Trigger)
	if err := d.selectionStillCurrent(dep); err != nil {
		appendLog("ERROR: release approval changed before checkout")
		d.finalizeDeploy(dep, "FAILED", startTime)
		return
	}

	statusCmd := exec.Command("git", "status", "--porcelain", "--untracked-files=no")
	statusCmd.Dir = d.repoDir
	if out, err := statusCmd.CombinedOutput(); err != nil || len(bytes.TrimSpace(out)) > 0 {
		appendLog("ERROR: deployment checkout is not clean: %s (%v)", strings.TrimSpace(string(out)), err)
		d.finalizeDeploy(dep, "FAILED", startTime)
		return
	}

	fetchCmd := exec.Command("git", "fetch", "origin", dep.Branch)
	fetchCmd.Dir = d.repoDir
	if out, err := fetchCmd.CombinedOutput(); err != nil {
		appendLog("ERROR: git fetch failed: %s (%v)", strings.TrimSpace(string(out)), err)
		d.finalizeDeploy(dep, "FAILED", startTime)
		return
	}

	ancestryCmd := exec.Command("git", "merge-base", "--is-ancestor", dep.Commit, "FETCH_HEAD")
	ancestryCmd.Dir = d.repoDir
	if err := ancestryCmd.Run(); err != nil {
		appendLog("ERROR: approved commit is not on the configured remote branch")
		d.finalizeDeploy(dep, "FAILED", startTime)
		return
	}
	if err := d.selectionStillCurrent(dep); err != nil {
		appendLog("ERROR: release approval changed during fetch")
		d.finalizeDeploy(dep, "FAILED", startTime)
		return
	}
	resetCmd := exec.Command("git", "reset", "--hard", dep.Commit)
	resetCmd.Dir = d.repoDir
	if out, err := resetCmd.CombinedOutput(); err != nil {
		appendLog("ERROR: checkout of CI-verified commit failed: %s (%v)", strings.TrimSpace(string(out)), err)
		d.finalizeDeploy(dep, "FAILED", startTime)
		return
	}

	deployScript := filepath.Join(d.repoDir, "deploy", "scripts", "deploy.sh")
	if _, err := os.Stat(deployScript); err != nil {
		appendLog("ERROR: deploy.sh is missing: %v", err)
		d.finalizeDeploy(dep, "FAILED", startTime)
		return
	}

	appendLog("Pulling and activating CI-built images")
	if err := d.selectionStillCurrent(dep); err != nil {
		appendLog("ERROR: release approval changed before activation")
		d.finalizeDeploy(dep, "FAILED", startTime)
		return
	}
	runCmd := exec.Command("bash", deployScript)
	runCmd.Dir = d.repoDir
	runCmd.Env = make([]string, 0, len(os.Environ())+1)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "DEPLOY_IMAGE_TAG=") {
			runCmd.Env = append(runCmd.Env, entry)
		}
	}
	runCmd.Env = append(runCmd.Env, "DEPLOY_IMAGE_TAG="+dep.Commit)
	out, err := runCmd.CombinedOutput()
	appendLog("%s", string(out))
	if err != nil {
		appendLog("ERROR: deploy.sh failed: %v", err)
		d.finalizeDeploy(dep, "FAILED", startTime)
		return
	}

	appendLog("Checking backend readiness")
	for _, probe := range []struct{ name, container, url string }{
		{"Core Backend", "hackways-core-backend", "http://localhost:8081/health"},
		{"Global Backend", "hackways-global-backend", "http://localhost:8082/health"},
	} {
		cmd := exec.Command("docker", "exec", probe.container, "wget", "-qO-", probe.url)
		probeOutput, probeErr := cmd.CombinedOutput()
		if probeErr != nil {
			appendLog("ERROR: %s health check failed: %v (%s)", probe.name, probeErr, strings.TrimSpace(string(probeOutput)))
			d.finalizeDeploy(dep, "FAILED", startTime)
			return
		}
		appendLog("%s health check passed", probe.name)
	}

	appendLog("Deployment succeeded: state = READY")
	d.finalizeDeploy(dep, "READY", startTime)
}
func (d *Deployer) finalizeDeploy(dep *Deployment, status string, startTime time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if status != "REFUSED" {
		if err := d.writeState(dep, status); err != nil {
			d.logger.Error("failed to persist release result; automatic deployment disabled")
			d.autoEnabled = false
			status = "FAILED"
		}
	}
	duration := time.Since(startTime)
	d.isDeploying = false
	d.active = nil

	d.logger.Info("deployment finished",
		"id", dep.ID,
		"status", dep.Status,
		"duration", duration.String(),
		"commit", dep.ShortSHA,
	)
	if d.pending != nil && d.autoEnabled {
		next := d.pending
		d.pending = nil
		d.isDeploying = true
		d.active = next
		go d.runAccepted(next)
	}
}

// HTTP Handlers

func (d *Deployer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	if !validateSignature(body, d.webhookSecret, r.Header.Get("X-Hub-Signature-256")) {
		http.Error(w, "Invalid webhook signature", http.StatusUnauthorized)
		return
	}
	event := r.Header.Get("X-GitHub-Event")
	if event == "ping" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
		return
	}
	if event != "workflow_run" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("event ignored"))
		return
	}
	var payload struct {
		Action     string `json:"action"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		WorkflowRun struct {
			Name       string `json:"name"`
			HeadBranch string `json:"head_branch"`
			HeadSHA    string `json:"head_sha"`
			Conclusion string `json:"conclusion"`
		} `json:"workflow_run"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid workflow payload", http.StatusBadRequest)
		return
	}
	run := payload.WorkflowRun
	sha, shaErr := hex.DecodeString(run.HeadSHA)
	if payload.Repository.FullName != d.repository || payload.Action != "completed" || run.Name != "Release Candidate" || run.Conclusion != "success" || run.HeadBranch != d.branch || shaErr != nil || len(sha) != 20 {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("workflow ignored"))
		return
	}
	dep, message, err := d.triggerDeploy("Release Candidate completed", run.HeadSHA)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		http.Error(w, "Release approval or deployment state unavailable", http.StatusServiceUnavailable)
		return
	}
	if dep == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(dep)
}
func (d *Deployer) handleHealth(w http.ResponseWriter, r *http.Request) {
	d.mu.Lock()
	defer d.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":                       "operational",
		"service":                      "Hackways CI deployment webhook",
		"is_deploying":                 d.isDeploying,
		"branch":                       d.branch,
		"automatic_deployment_enabled": d.autoEnabled,
	})
}

func validateSignature(payload []byte, secret, signatureHeader string) bool {
	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}
	expectedSig := signatureHeader[7:]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	computedSig := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedSig), []byte(computedSig))
}
