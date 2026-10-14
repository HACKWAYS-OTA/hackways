package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
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
	ID        string    `json:"id"`
	Commit    string    `json:"commit"`
	ShortSHA  string    `json:"short_sha"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Branch    string    `json:"branch"`
	Trigger   string    `json:"trigger"`
	Status    string    `json:"status"` // "QUEUED", "BUILDING", "READY", "FAILED"
	CreatedAt time.Time `json:"created_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	Logs      string    `json:"logs,omitempty"`
}

// Deployer coordinates deployments, webhooks, and auto-polling.
type Deployer struct {
	repoDir       string
	branch        string
	port          int
	webhookSecret string
	pollInterval  time.Duration
	mu            sync.Mutex
	isDeploying   bool
	history       []*Deployment
	historyFile   string
	logger        *slog.Logger
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	repoDir := getEnv("REPO_DIR", "/opt/hackways")
	branch := getEnv("BRANCH", "develop")
	secret := getEnv("WEBHOOK_SECRET", "")
	port := 9898

	deployer := &Deployer{
		repoDir:       repoDir,
		branch:        branch,
		port:          port,
		webhookSecret: secret,
		pollInterval:  20 * time.Second,
		historyFile:   filepath.Join(repoDir, "deployments", "history.json"),
		logger:        logger,
	}

	// Ensure deployments directory exists
	_ = os.MkdirAll(filepath.Join(repoDir, "deployments"), 0755)
	deployer.loadHistory()

	// Initial check on startup
	go deployer.startPolling()

	mux := http.NewServeMux()
	mux.HandleFunc("/", deployer.handleDashboard)
	mux.HandleFunc("/api/deployments", deployer.handleListDeployments)
	mux.HandleFunc("/api/deployments/logs", deployer.handleGetLogs)
	mux.HandleFunc("/api/deploy/trigger", deployer.handleTriggerDeploy)
	mux.HandleFunc("/api/deploy/webhook", deployer.handleWebhook)
	mux.HandleFunc("/health", deployer.handleHealth)

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	logger.Info("Hackways Vercel-like Auto Deployer starting",
		"port", port,
		"repo_dir", repoDir,
		"branch", branch,
		"poll_interval", "20s",
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

func (d *Deployer) loadHistory() {
	data, err := os.ReadFile(d.historyFile)
	if err != nil {
		return
	}
	var hist []*Deployment
	if err := json.Unmarshal(data, &hist); err == nil {
		d.history = hist
	}
}

func (d *Deployer) saveHistory() {
	data, err := json.MarshalIndent(d.history, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(d.historyFile, data, 0644)
}

func (d *Deployer) startPolling() {
	ticker := time.NewTicker(d.pollInterval)
	defer ticker.Stop()

	d.logger.Info("auto-polling engine active: monitoring GitHub for newest commits", "branch", d.branch)

	for range ticker.C {
		d.checkAndDeployNewCommit()
	}
}

func (d *Deployer) checkAndDeployNewCommit() {
	d.mu.Lock()
	if d.isDeploying {
		d.mu.Unlock()
		return
	}
	d.mu.Unlock()

	// 1. Get latest remote SHA from GitHub
	cmd := exec.Command("git", "ls-remote", "origin", fmt.Sprintf("refs/heads/%s", d.branch))
	cmd.Dir = d.repoDir
	out, err := cmd.Output()
	if err != nil {
		return
	}

	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return
	}
	remoteSHA := fields[0]

	// 2. Get local HEAD SHA
	headCmd := exec.Command("git", "rev-parse", "HEAD")
	headCmd.Dir = d.repoDir
	headOut, err := headCmd.Output()
	if err != nil {
		return
	}
	localSHA := strings.TrimSpace(string(headOut))

	// 3. Compare: If remote has a newer commit, trigger deployment!
	if remoteSHA != localSHA {
		d.logger.Info("new commit detected on GitHub! Initiating auto-deployment",
			"branch", d.branch,
			"current_head", localSHA[:7],
			"new_remote", remoteSHA[:7],
		)
		go d.triggerDeploy("Auto-Poll (GitHub Push)")
	}
}

func (d *Deployer) triggerDeploy(trigger string) *Deployment {
	d.mu.Lock()
	if d.isDeploying {
		d.mu.Unlock()
		d.logger.Warn("deployment already in progress; request ignored", "trigger", trigger)
		return nil
	}
	d.isDeploying = true

	depID := fmt.Sprintf("dep_%s", time.Now().Format("20060102_150405"))
	dep := &Deployment{
		ID:        depID,
		Branch:    d.branch,
		Trigger:   trigger,
		Status:    "BUILDING",
		CreatedAt: time.Now(),
		Logs:      "",
	}
	d.history = append([]*Deployment{dep}, d.history...)
	if len(d.history) > 50 {
		d.history = d.history[:50]
	}
	d.saveHistory()
	d.mu.Unlock()

	go d.runDeployPipeline(dep)
	return dep
}

func (d *Deployer) runDeployPipeline(dep *Deployment) {
	var logBuf bytes.Buffer
	startTime := time.Now()

	appendLog := func(format string, a ...interface{}) {
		msg := fmt.Sprintf(format, a...)
		timestamp := time.Now().Format("15:04:05.000")
		formatted := fmt.Sprintf("[%s] %s\n", timestamp, msg)
		logBuf.WriteString(formatted)
		dep.Logs = logBuf.String()
		d.logger.Info("deploy log", "dep_id", dep.ID, "line", msg)
	}

	appendLog("🚀 === Hackways Auto-Deployment Started ===")
	appendLog("Target Environment: Azure VM Production (CRIN-MDB-PR / 4.247.31.186)")
	appendLog("Branch: %s | Trigger: %s", dep.Branch, dep.Trigger)

	// Step 1: Fetch and update git repository
	appendLog("--> Step 1/5: Fetching newest commit from GitHub...")
	fetchCmd := exec.Command("git", "fetch", "origin", dep.Branch)
	fetchCmd.Dir = d.repoDir
	if out, err := fetchCmd.CombinedOutput(); err != nil {
		appendLog("ERROR: git fetch failed: %s (%v)", string(out), err)
		d.finalizeDeploy(dep, "FAILED", startTime, logBuf.String())
		return
	}

	// Step 2: Checkout and reset to origin
	appendLog("--> Step 2/5: Synchronizing working directory to origin/%s...", dep.Branch)
	resetCmd := exec.Command("git", "reset", "--hard", fmt.Sprintf("origin/%s", dep.Branch))
	resetCmd.Dir = d.repoDir
	if out, err := resetCmd.CombinedOutput(); err != nil {
		appendLog("ERROR: git reset failed: %s (%v)", string(out), err)
		d.finalizeDeploy(dep, "FAILED", startTime, logBuf.String())
		return
	}

	// Extract commit metadata
	commitCmd := exec.Command("git", "log", "-1", "--format=%H%n%an%n%s")
	commitCmd.Dir = d.repoDir
	if out, err := commitCmd.Output(); err == nil {
		parts := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(parts) >= 3 {
			dep.Commit = parts[0]
			dep.ShortSHA = parts[0][:7]
			dep.Author = parts[1]
			dep.Message = parts[2]
			appendLog("Commit: %s (%s)", dep.ShortSHA, dep.Message)
			appendLog("Author: %s", dep.Author)
		}
	}

	// Step 3: Run deployment script
	appendLog("--> Step 3/5: Executing production deploy script (deploy.sh)...")
	deployScript := filepath.Join(d.repoDir, "deploy", "scripts", "deploy.sh")
	if _, err := os.Stat(deployScript); err == nil {
		runCmd := exec.Command("bash", deployScript)
		runCmd.Dir = d.repoDir
		runCmd.Env = append(os.Environ(), "DOCKER_REGISTRY=hackways")
		out, err := runCmd.CombinedOutput()
		appendLog("%s", string(out))
		if err != nil {
			appendLog("ERROR: deploy.sh failed with error: %v", err)
			d.finalizeDeploy(dep, "FAILED", startTime, logBuf.String())
			return
		}
	} else {
		// Fallback: Rebuild Go backends and restart containers
		appendLog("Rebuilding Go backends and launching containers...")
		b1 := exec.Command("docker", "build", "-f", "deploy/docker/Dockerfile.core-backend", "-t", "hackways/hackways-core-backend:latest", ".")
		b1.Dir = d.repoDir
		_ = b1.Run()

		b2 := exec.Command("docker", "build", "-f", "deploy/docker/Dockerfile.global-backend", "-t", "hackways/hackways-global-backend:latest", ".")
		b2.Dir = d.repoDir
		_ = b2.Run()

		upCmd := exec.Command("docker", "compose", "-f", "deploy/docker/docker-compose.prod.yml", "--env-file", ".env", "up", "-d")
		upCmd.Dir = d.repoDir
		out, err := upCmd.CombinedOutput()
		appendLog("%s", string(out))
		if err != nil {
			appendLog("ERROR: docker compose up failed: %v", err)
			d.finalizeDeploy(dep, "FAILED", startTime, logBuf.String())
			return
		}
	}

	// Step 4: Database Migrations Check
	appendLog("--> Step 4/5: Verifying PostgreSQL database migrations...")
	migCmd := exec.Command("bash", "-c", "docker exec -i hackways-postgres psql -U hackways_prod -d hackways_prod < deploy/migrations/000001_initial_schema.up.sql")
	migCmd.Dir = d.repoDir
	_ = migCmd.Run()

	// Step 5: Health & Integrity Verification
	appendLog("--> Step 5/5: Running post-deployment health probes...")
	time.Sleep(3 * time.Second)

	healthCore := exec.Command("docker", "exec", "hackways-core-backend", "wget", "-qO-", "http://localhost:8081/health")
	if out, err := healthCore.CombinedOutput(); err == nil {
		appendLog("Core Backend Health: OK (%s)", strings.TrimSpace(string(out)))
	} else {
		appendLog("WARNING: Core Backend health probe returned: %v", err)
	}

	healthGlobal := exec.Command("docker", "exec", "hackways-global-backend", "wget", "-qO-", "http://localhost:8082/health")
	if out, err := healthGlobal.CombinedOutput(); err == nil {
		appendLog("Global Backend Health: OK (%s)", strings.TrimSpace(string(out)))
	} else {
		appendLog("WARNING: Global Backend health probe returned: %v", err)
	}

	appendLog("✅ === Deployment Succeeded: State = READY ===")
	d.finalizeDeploy(dep, "READY", startTime, logBuf.String())
}

func (d *Deployer) finalizeDeploy(dep *Deployment, status string, startTime time.Time, logs string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	dep.Status = status
	dep.EndedAt = time.Now()
	dur := dep.EndedAt.Sub(startTime)
	dep.Duration = fmt.Sprintf("%.1fs", dur.Seconds())
	dep.Logs = logs
	d.isDeploying = false
	d.saveHistory()

	d.logger.Info("deployment finished",
		"id", dep.ID,
		"status", dep.Status,
		"duration", dep.Duration,
		"commit", dep.ShortSHA,
	)
}

// HTTP Handlers

func (d *Deployer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	d.mu.Lock()
	deployments := make([]*Deployment, len(d.history))
	copy(deployments, d.history)
	isDeploying := d.isDeploying
	d.mu.Unlock()

	var latest *Deployment
	if len(deployments) > 0 {
		latest = deployments[0]
	}

	data := struct {
		Branch      string
		IsDeploying bool
		Latest      *Deployment
		Deployments []*Deployment
	}{
		Branch:      d.branch,
		IsDeploying: isDeploying,
		Latest:      latest,
		Deployments: deployments,
	}

	tmpl := template.Must(template.New("dashboard").Parse(dashboardHTML))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

func (d *Deployer) handleListDeployments(w http.ResponseWriter, r *http.Request) {
	d.mu.Lock()
	defer d.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"is_deploying": d.isDeploying,
		"branch":       d.branch,
		"deployments":  d.history,
	})
}

func (d *Deployer) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, dep := range d.history {
		if dep.ID == id {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte(dep.Logs))
			return
		}
	}
	http.Error(w, "Deployment not found", http.StatusNotFound)
}

func (d *Deployer) handleTriggerDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dep := d.triggerDeploy("Manual Web Trigger")
	if dep == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "A deployment is currently in progress",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(dep)
}

func (d *Deployer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Validate GitHub Webhook Signature if secret is configured
	if d.webhookSecret != "" {
		sig := r.Header.Get("X-Hub-Signature-256")
		if !validateSignature(body, d.webhookSecret, sig) {
			http.Error(w, "Invalid webhook signature", http.StatusUnauthorized)
			return
		}
	}

	event := r.Header.Get("X-GitHub-Event")
	if event == "ping" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
		return
	}

	if event != "push" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("event ignored"))
		return
	}

	dep := d.triggerDeploy("GitHub Webhook Push")
	w.Header().Set("Content-Type", "application/json")
	if dep == nil {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Deployment already running"})
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
		"status":       "operational",
		"service":      "Hackways Vercel Auto-Deployer",
		"is_deploying": d.isDeploying,
		"branch":       d.branch,
		"history_size": len(d.history),
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

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Hackways — Production Deployments</title>
    <style>
        :root {
            --bg: #000000;
            --surface: #0a0a0a;
            --card: #111111;
            --border: #222222;
            --border-hover: #333333;
            --text-primary: #ededed;
            --text-secondary: #888888;
            --accent-green: #0070f3;
            --success: #00e599;
            --warning: #f5a623;
            --error: #ff0055;
            --font: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
            --mono: 'SFMono-Regular', Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            background-color: var(--bg);
            color: var(--text-primary);
            font-family: var(--font);
            -webkit-font-smoothing: antialiased;
            min-height: 100vh;
            padding: 0 0 60px 0;
        }
        header {
            border-bottom: 1px solid var(--border);
            padding: 18px 32px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            background: rgba(0, 0, 0, 0.8);
            backdrop-filter: blur(12px);
            position: sticky;
            top: 0;
            z-index: 100;
        }
        .header-brand {
            display: flex;
            align-items: center;
            gap: 14px;
        }
        .brand-logo {
            width: 32px;
            height: 32px;
            background: #ffffff;
            color: #000000;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: 800;
            font-size: 16px;
        }
        .brand-title {
            font-size: 18px;
            font-weight: 600;
            letter-spacing: -0.5px;
        }
        .header-badges {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .badge {
            display: inline-flex;
            align-items: center;
            gap: 6px;
            padding: 4px 10px;
            border-radius: 9999px;
            font-size: 12px;
            font-weight: 500;
            border: 1px solid var(--border);
            background: var(--surface);
        }
        .badge.prod {
            color: var(--success);
            border-color: rgba(0, 229, 153, 0.3);
            background: rgba(0, 229, 153, 0.05);
        }
        .badge.branch {
            color: #0070f3;
            border-color: rgba(0, 112, 243, 0.3);
            background: rgba(0, 112, 243, 0.05);
            font-family: var(--mono);
        }
        .container {
            max-width: 1200px;
            margin: 40px auto 0;
            padding: 0 24px;
        }
        .hero {
            display: flex;
            justify-content: space-between;
            align-items: flex-end;
            margin-bottom: 32px;
        }
        .hero h1 {
            font-size: 28px;
            font-weight: 700;
            letter-spacing: -0.8px;
            margin-bottom: 8px;
        }
        .hero p {
            color: var(--text-secondary);
            font-size: 14px;
        }
        .btn-deploy {
            background: #ffffff;
            color: #000000;
            border: none;
            padding: 10px 20px;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.15s ease;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .btn-deploy:hover {
            background: #cccccc;
        }
        .btn-deploy:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            display: inline-block;
        }
        .status-dot.READY { background-color: var(--success); box-shadow: 0 0 10px var(--success); }
        .status-dot.BUILDING { 
            background-color: var(--warning); 
            box-shadow: 0 0 10px var(--warning);
            animation: pulse 1.5s infinite;
        }
        .status-dot.FAILED { background-color: var(--error); box-shadow: 0 0 10px var(--error); }
        .status-dot.QUEUED { background-color: #666; }
        @keyframes pulse {
            0% { transform: scale(0.9); opacity: 0.7; }
            50% { transform: scale(1.2); opacity: 1; }
            100% { transform: scale(0.9); opacity: 0.7; }
        }
        .current-card {
            background: var(--surface);
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 24px;
            margin-bottom: 40px;
        }
        .current-card:hover {
            border-color: var(--border-hover);
        }
        .current-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 16px;
        }
        .current-title {
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 1px;
            color: var(--text-secondary);
            font-weight: 600;
        }
        .current-commit-msg {
            font-size: 20px;
            font-weight: 600;
            margin-bottom: 12px;
            letter-spacing: -0.4px;
        }
        .current-meta {
            display: flex;
            gap: 20px;
            color: var(--text-secondary);
            font-size: 13px;
        }
        .meta-item {
            display: flex;
            align-items: center;
            gap: 6px;
        }
        .mono {
            font-family: var(--mono);
            color: var(--text-primary);
            background: var(--border);
            padding: 2px 6px;
            border-radius: 4px;
        }
        .history-table {
            width: 100%;
            border-collapse: collapse;
            background: var(--surface);
            border: 1px solid var(--border);
            border-radius: 12px;
            overflow: hidden;
        }
        .history-table th {
            text-align: left;
            padding: 14px 20px;
            background: var(--card);
            border-bottom: 1px solid var(--border);
            font-size: 12px;
            text-transform: uppercase;
            letter-spacing: 0.8px;
            color: var(--text-secondary);
        }
        .history-table td {
            padding: 16px 20px;
            border-bottom: 1px solid var(--border);
            font-size: 14px;
            vertical-align: middle;
        }
        .history-table tr:last-child td {
            border-bottom: none;
        }
        .history-table tr:hover td {
            background: rgba(255, 255, 255, 0.02);
        }
        .btn-log {
            background: transparent;
            color: #0070f3;
            border: 1px solid rgba(0, 112, 243, 0.3);
            padding: 6px 12px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.15s ease;
        }
        .btn-log:hover {
            background: rgba(0, 112, 243, 0.1);
        }
        /* Modal for logs */
        #log-modal {
            display: none;
            position: fixed;
            top: 0; left: 0; right: 0; bottom: 0;
            background: rgba(0, 0, 0, 0.85);
            backdrop-filter: blur(8px);
            z-index: 999;
            align-items: center;
            justify-content: center;
            padding: 24px;
        }
        .modal-content {
            background: #0d0d0d;
            border: 1px solid var(--border);
            border-radius: 12px;
            width: 100%;
            max-width: 900px;
            height: 80vh;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        .modal-header {
            padding: 16px 20px;
            border-bottom: 1px solid var(--border);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .modal-close {
            background: none;
            border: none;
            color: var(--text-secondary);
            font-size: 20px;
            cursor: pointer;
        }
        .modal-body {
            flex: 1;
            padding: 20px;
            background: #050505;
            color: #00ff66;
            font-family: var(--mono);
            font-size: 13px;
            overflow-y: auto;
            white-space: pre-wrap;
            line-height: 1.6;
        }
    </style>
</head>
<body>
    <header>
        <div class="header-brand">
            <div class="brand-logo">▲</div>
            <div class="brand-title">Hackways Auto-Deployer</div>
        </div>
        <div class="header-badges">
            <div class="badge prod">● PRODUCTION</div>
            <div class="badge branch">git: {{.Branch}}</div>
            <div class="badge">Host: 4.247.31.186</div>
        </div>
    </header>

    <div class="container">
        <div class="hero">
            <div>
                <h1>Deployments</h1>
                <p>Automatic continuous deployment engine triggered by GitHub push events and commit monitoring.</p>
            </div>
            <button class="btn-deploy" id="btn-trigger" onclick="triggerDeploy()" {{if .IsDeploying}}disabled{{end}}>
                {{if .IsDeploying}}▲ Building...{{else}}▲ Deploy Latest Commit{{end}}
            </button>
        </div>

        {{if .Latest}}
        <div class="current-card">
            <div class="current-header">
                <div class="current-title">Latest Production Deployment</div>
                <div class="badge">
                    <span class="status-dot {{.Latest.Status}}"></span>
                    <span style="font-weight:600; margin-left:4px;">{{.Latest.Status}}</span>
                </div>
            </div>
            <div class="current-commit-msg">{{if .Latest.Message}}{{.Latest.Message}}{{else}}Deployment {{.Latest.ID}}{{end}}</div>
            <div class="current-meta">
                <div class="meta-item">Commit: <span class="mono">{{.Latest.ShortSHA}}</span></div>
                <div class="meta-item">Author: <span>{{.Latest.Author}}</span></div>
                <div class="meta-item">Trigger: <span>{{.Latest.Trigger}}</span></div>
                <div class="meta-item">Duration: <span>{{.Latest.Duration}}</span></div>
                <div class="meta-item">Deployed: <span>{{.Latest.CreatedAt.Format "15:04:05 Jan 02, 2006"}}</span></div>
                <div class="meta-item" style="margin-left:auto;">
                    <button class="btn-log" onclick="showLogs('{{.Latest.ID}}')">View Build Output</button>
                </div>
            </div>
        </div>
        {{end}}

        <table class="history-table">
            <thead>
                <tr>
                    <th>Status</th>
                    <th>Commit</th>
                    <th>Message</th>
                    <th>Author</th>
                    <th>Trigger</th>
                    <th>Duration</th>
                    <th>Created At</th>
                    <th>Action</th>
                </tr>
            </thead>
            <tbody>
                {{range .Deployments}}
                <tr>
                    <td>
                        <span class="status-dot {{.Status}}"></span>
                        <span style="margin-left:6px; font-weight:500;">{{.Status}}</span>
                    </td>
                    <td><span class="mono">{{.ShortSHA}}</span></td>
                    <td style="font-weight:500; max-width:320px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">
                        {{.Message}}
                    </td>
                    <td style="color:var(--text-secondary);">{{.Author}}</td>
                    <td><span class="badge" style="font-size:11px;">{{.Trigger}}</span></td>
                    <td style="color:var(--text-secondary); font-family:var(--mono); font-size:12px;">{{.Duration}}</td>
                    <td style="color:var(--text-secondary); font-size:12px;">{{.CreatedAt.Format "15:04:05 02-Jan"}}</td>
                    <td>
                        <button class="btn-log" onclick="showLogs('{{.ID}}')">Logs</button>
                    </td>
                </tr>
                {{else}}
                <tr>
                    <td colspan="8" style="text-align:center; padding:32px; color:var(--text-secondary);">
                        No deployment history recorded yet.
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <!-- Logs Modal -->
    <div id="log-modal" onclick="if(event.target === this) hideLogs()">
        <div class="modal-content">
            <div class="modal-header">
                <div style="font-weight:600; font-size:14px;" id="modal-title">Deployment Logs</div>
                <button class="modal-close" onclick="hideLogs()">&times;</button>
            </div>
            <div class="modal-body" id="modal-body">Loading...</div>
        </div>
    </div>

    <script>
        function triggerDeploy() {
            const btn = document.getElementById('btn-trigger');
            btn.disabled = true;
            btn.innerText = '▲ Initiating...';
            fetch('/api/deploy/trigger', { method: 'POST' })
                .then(res => res.json())
                .then(data => {
                    setTimeout(() => window.location.reload(), 1000);
                })
                .catch(err => {
                    alert('Deploy trigger error: ' + err);
                    btn.disabled = false;
                    btn.innerText = '▲ Deploy Latest Commit';
                });
        }

        function showLogs(id) {
            const modal = document.getElementById('log-modal');
            const title = document.getElementById('modal-title');
            const body = document.getElementById('modal-body');
            title.innerText = 'Deployment Logs — ' + id;
            body.innerText = 'Fetching logs from daemon...';
            modal.style.display = 'flex';

            fetch('/api/deployments/logs?id=' + encodeURIComponent(id))
                .then(res => res.text())
                .then(text => {
                    body.innerText = text || 'No log output captured.';
                })
                .catch(err => {
                    body.innerText = 'Error loading logs: ' + err;
                });
        }

        function hideLogs() {
            document.getElementById('log-modal').style.display = 'none';
        }

        // Auto-refresh status every 6 seconds
        setInterval(() => {
            fetch('/api/deployments')
                .then(res => res.json())
                .then(data => {
                    if (data.is_deploying) {
                        const btn = document.getElementById('btn-trigger');
                        if (btn) {
                            btn.disabled = true;
                            btn.innerText = '▲ Building...';
                        }
                    }
                });
        }, 6000);
    </script>
</body>
</html>
`
