// Package zellij wraps the `zellij action` CLI used by worker.
package zellij

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InSession reports whether the process is running inside a zellij session.
func InSession() bool {
	return os.Getenv("ZELLIJ") != ""
}

type tab struct {
	Name string `json:"name"`
}

// TabExists reports whether a tab with the given name exists in the current session.
func TabExists(name string) (bool, error) {
	cmd := exec.Command("zellij", "action", "list-tabs", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("zellij action list-tabs: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var tabs []tab
	if err := json.Unmarshal(stdout.Bytes(), &tabs); err != nil {
		return false, fmt.Errorf("parse list-tabs json: %w", err)
	}
	for _, t := range tabs {
		if t.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// FocusTab switches focus to the named tab.
func FocusTab(name string) error {
	return runAction("go-to-tab-name", name)
}

// NewTab creates a new tab from a raw KDL layout string, with the given name and cwd.
func NewTab(name, cwd, layoutKDL string) error {
	cmd := exec.Command("zellij", "action", "new-tab",
		"--name", name,
		"--cwd", cwd,
		"--layout-string", layoutKDL,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zellij action new-tab: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func runAction(args ...string) error {
	full := append([]string{"action"}, args...)
	cmd := exec.Command("zellij", full...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("zellij %s: %w: %s", strings.Join(full, " "), err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
