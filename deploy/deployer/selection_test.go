package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testDeployer(t *testing.T) *Deployer {
	t.Helper()
	dir := t.TempDir()
	return &Deployer{branch: "main", repository: "HACKWAYS-OTA/hackways", webhookSecret: "isolated-webhook-test-key", autoEnabled: true, approvalFile: filepath.Join(dir, "approval.json"), stateFile: filepath.Join(dir, "state.json"), logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func approve(t *testing.T, d *Deployer, generation uint64, sha string) {
	t.Helper()
	data, err := json.Marshal(ReleaseSelection{d.repository, d.branch, sha, generation})
	if err != nil {
		t.Fatal(err)
	}
	temp := d.approvalFile + ".next"
	if err = os.WriteFile(temp, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(temp, d.approvalFile); err != nil {
		t.Fatal(err)
	}
}

func deliver(d *Deployer, sha string) *httptest.ResponseRecorder {
	body := []byte(`{"action":"completed","repository":{"full_name":"HACKWAYS-OTA/hackways"},"workflow_run":{"name":"Release Candidate","head_branch":"main","head_sha":"` + sha + `","conclusion":"success"}}`)
	mac := hmac.New(sha256.New, []byte(d.webhookSecret))
	_, _ = mac.Write(body)
	r := httptest.NewRequest(http.MethodPost, "/api/deploy/webhook", bytes.NewReader(body))
	r.Header.Set("X-GitHub-Event", "workflow_run")
	r.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	w := httptest.NewRecorder()
	d.handleWebhook(w, r)
	return w
}

func idle(t *testing.T, d *Deployer) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		d.mu.Lock()
		busy := d.isDeploying
		d.mu.Unlock()
		if !busy {
			return
		}
		select {
		case <-deadline:
			t.Fatal("deployment worker did not settle")
		case <-time.After(time.Millisecond):
		}
	}
}

func receive(t *testing.T, ch <-chan uint64, expected uint64) {
	t.Helper()
	select {
	case actual := <-ch:
		if actual != expected {
			t.Fatalf("deployed generation %d, expected %d", actual, expected)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("accepted deployment was lost")
	}
}

func TestApprovedReleaseRejectsStaleAndReplayedEventsAcrossRestart(t *testing.T) {
	d := testDeployer(t)
	oldSHA, currentSHA := strings.Repeat("a", 40), strings.Repeat("b", 40)
	var calls atomic.Int32
	d.pipeline = func(*Deployment) error { calls.Add(1); return nil }
	approve(t, d, 2, currentSHA)
	if w := deliver(d, oldSHA); w.Code != 200 {
		t.Fatalf("out-of-order event status %d", w.Code)
	}
	if w := deliver(d, currentSHA); w.Code != 202 {
		t.Fatalf("approved event status %d: %s", w.Code, w.Body.String())
	}
	idle(t, d)
	if w := deliver(d, currentSHA); w.Code != 200 {
		t.Fatalf("repeated event status %d", w.Code)
	}
	approve(t, d, 1, oldSHA)
	if w := deliver(d, oldSHA); w.Code != 200 {
		t.Fatalf("stale approval status %d", w.Code)
	}
	restarted := testDeployer(t)
	restarted.approvalFile, restarted.stateFile, restarted.pipeline = d.approvalFile, d.stateFile, d.pipeline
	approve(t, restarted, 2, currentSHA)
	if w := deliver(restarted, currentSHA); w.Code != 200 {
		t.Fatalf("restart replay status %d", w.Code)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one activation, got %d", calls.Load())
	}
}

func TestConcurrentBusyDeliveriesQueueApprovedReleaseOnce(t *testing.T) {
	d := testDeployer(t)
	firstSHA, nextSHA := strings.Repeat("a", 40), strings.Repeat("b", 40)
	started := make(chan uint64, 4)
	releaseFirst := make(chan struct{})
	d.pipeline = func(dep *Deployment) error {
		started <- dep.Generation
		if dep.Generation == 1 {
			<-releaseFirst
		}
		return nil
	}
	approve(t, d, 1, firstSHA)
	if w := deliver(d, firstSHA); w.Code != 202 {
		t.Fatal(w.Code)
	}
	receive(t, started, 1)
	approve(t, d, 2, nextSHA)
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if w := deliver(d, nextSHA); w.Code != 202 {
				t.Errorf("busy valid event dropped: HTTP %d", w.Code)
			}
		}()
	}
	wg.Wait()
	if w := deliver(d, firstSHA); w.Code != 200 {
		t.Errorf("stale event while busy: HTTP %d", w.Code)
	}
	close(releaseFirst)
	receive(t, started, 2)
	idle(t, d)
	if len(started) != 0 {
		t.Fatal("duplicate activation")
	}
	state, err := d.readState()
	if err != nil || state.Generation != 2 || state.Commit != nextSHA || state.Status != "READY" {
		t.Fatalf("wrong durable selection: %+v %v", state, err)
	}
}

func TestSupersededQueuedReleaseDoesNotActivate(t *testing.T) {
	d := testDeployer(t)
	started := make(chan uint64, 4)
	release := make(chan struct{})
	d.pipeline = func(dep *Deployment) error {
		started <- dep.Generation
		if dep.Generation == 1 {
			<-release
		}
		return nil
	}
	approve(t, d, 1, strings.Repeat("a", 40))
	deliver(d, strings.Repeat("a", 40))
	receive(t, started, 1)
	approve(t, d, 2, strings.Repeat("b", 40))
	deliver(d, strings.Repeat("b", 40))
	approve(t, d, 3, strings.Repeat("c", 40))
	deliver(d, strings.Repeat("c", 40))
	close(release)
	receive(t, started, 3)
	idle(t, d)
	if len(started) != 0 {
		t.Fatal("superseded release activated")
	}
}

func TestDisabledOrMissingApprovalCannotDeploy(t *testing.T) {
	d := testDeployer(t)
	d.autoEnabled = false
	d.pipeline = func(*Deployment) error { t.Error("disabled activation"); return nil }
	if w := deliver(d, strings.Repeat("a", 40)); w.Code != 200 {
		t.Fatal(w.Code)
	}
	d.autoEnabled = true
	if w := deliver(d, strings.Repeat("a", 40)); w.Code != 503 {
		t.Fatal(w.Code)
	}
}

func TestFailedAttemptNeedsNewApprovalGeneration(t *testing.T) {
	d := testDeployer(t)
	var calls atomic.Int32
	d.pipeline = func(*Deployment) error { calls.Add(1); return errors.New("isolated failure") }
	sha := strings.Repeat("a", 40)
	approve(t, d, 1, sha)
	deliver(d, sha)
	idle(t, d)
	if w := deliver(d, sha); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if calls.Load() != 1 {
		t.Fatal("failed event replayed")
	}
	approve(t, d, 2, sha)
	deliver(d, sha)
	idle(t, d)
	if calls.Load() != 2 {
		t.Fatal("new operator approval not accepted")
	}
}
