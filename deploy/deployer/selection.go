package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReleaseSelection is an operator approval outside the Git checkout. Every
// attempt, including an intentional rollback, requires a new generation.
type ReleaseSelection struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	Generation uint64 `json:"generation"`
}

type releaseState struct {
	ReleaseSelection
	Status string `json:"status"`
}

func (d *Deployer) approvedRelease() (ReleaseSelection, error) {
	var selected ReleaseSelection
	data, err := os.ReadFile(d.approvalFile)
	if err != nil || json.Unmarshal(data, &selected) != nil {
		return selected, errors.New("approved release is unavailable or invalid")
	}
	sha, err := hex.DecodeString(selected.Commit)
	if err != nil || len(sha) != 20 || selected.Commit != strings.ToLower(selected.Commit) || selected.Generation == 0 || selected.Repository != d.repository || selected.Branch != d.branch {
		return selected, errors.New("approved release does not match configured repository and branch")
	}
	return selected, nil
}

func (d *Deployer) readState() (releaseState, error) {
	var state releaseState
	data, err := os.ReadFile(d.stateFile)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil || json.Unmarshal(data, &state) != nil {
		return state, errors.New("deployment state is unavailable or invalid")
	}
	if state.Generation == 0 || state.Repository != d.repository || state.Branch != d.branch || (state.Status != "BUILDING" && state.Status != "READY" && state.Status != "FAILED") {
		return state, errors.New("deployment state requires operator reconciliation")
	}
	return state, nil
}

// Persist before checkout/activation, so a restart cannot forget an attempt and
// permit an older event to roll back the release. State contains no credentials.
func (d *Deployer) writeState(dep *Deployment, status string) error {
	state := releaseState{ReleaseSelection: ReleaseSelection{d.repository, dep.Branch, dep.Commit, dep.Generation}, Status: status}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	dir := filepath.Dir(d.stateFile)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".release-state-")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(name, d.stateFile)
}

func (d *Deployer) selectionStillCurrent(dep *Deployment) error {
	selected, err := d.approvedRelease()
	if err != nil {
		return err
	}
	if selected.Commit != dep.Commit || selected.Generation != dep.Generation {
		return errors.New("release approval changed; stale activation refused")
	}
	return nil
}

func (d *Deployer) triggerDeploy(trigger, commit string) (*Deployment, string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.autoEnabled {
		return nil, "automatic deployment disabled", nil
	}
	selected, err := d.approvedRelease()
	if err != nil {
		return nil, "", err
	}
	if commit != selected.Commit {
		return nil, "unapproved or stale workflow ignored", nil
	}
	state, err := d.readState()
	if err != nil {
		return nil, "", err
	}
	if selected.Generation <= state.Generation {
		return nil, "release generation already attempted; replay ignored", nil
	}
	if d.active != nil && selected.Generation <= d.active.Generation {
		return d.active, "release already accepted", nil
	}
	if d.pending != nil && selected.Generation <= d.pending.Generation {
		return d.pending, "release already queued", nil
	}
	dep := &Deployment{
		ID: fmt.Sprintf("dep_%d_%s", selected.Generation, shortSHA(commit)), Commit: commit,
		ShortSHA: shortSHA(commit), Branch: d.branch, Generation: selected.Generation,
		Trigger: trigger, Status: "QUEUED", CreatedAt: time.Now(),
	}
	if d.isDeploying {
		d.pending = dep
		return dep, "approved release queued", nil
	}
	d.isDeploying = true
	d.active = dep
	go d.runAccepted(dep)
	return dep, "approved release accepted", nil
}

func (d *Deployer) runAccepted(dep *Deployment) {
	d.mu.Lock()
	err := d.selectionStillCurrent(dep)
	if err == nil {
		var state releaseState
		state, err = d.readState()
		if err == nil && dep.Generation <= state.Generation {
			err = errors.New("release generation already attempted")
		}
	}
	if err == nil {
		err = d.writeState(dep, "BUILDING")
	}
	d.mu.Unlock()
	if err != nil {
		d.logger.Error("deployment refused before checkout", "error", err)
		d.finalizeDeploy(dep, "REFUSED", dep.CreatedAt)
		return
	}
	if d.pipeline != nil {
		if err := d.pipeline(dep); err != nil {
			d.finalizeDeploy(dep, "FAILED", dep.CreatedAt)
		} else {
			d.finalizeDeploy(dep, "READY", dep.CreatedAt)
		}
		return
	}
	d.runDeployPipeline(dep)
}
