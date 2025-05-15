package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// runGHCommand executes a GitHub CLI command and returns its output
func runGHCommand(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running GitHub CLI command: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
