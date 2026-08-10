package script

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Runner executes shell scripts
type Runner struct {
	Timeout time.Duration
}

// NewRunner creates a new script runner
func NewRunner(timeout time.Duration) *Runner {
	return &Runner{Timeout: timeout}
}

// Run executes a shell script and returns an error if it fails
func (r *Runner) Run(scriptPath, workDir string) error {
	_, _, exitCode, err := r.RunWithDetails(scriptPath, workDir)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("script exited with code %d", exitCode)
	}
	return nil
}

// RunWithOutput executes a shell script and returns its stdout
func (r *Runner) RunWithOutput(scriptPath, workDir string) (string, error) {
	stdout, _, exitCode, err := r.RunWithDetails(scriptPath, workDir)
	if err != nil {
		return stdout, err
	}
	if exitCode != 0 {
		return stdout, fmt.Errorf("script exited with code %d", exitCode)
	}
	return stdout, nil
}

// RunWithDetails executes a shell script and returns stdout, stderr, exit code, and error
func (r *Runner) RunWithDetails(scriptPath, workDir string) (stdout, stderr string, exitCode int, err error) {
	// Validate script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", "", -1, fmt.Errorf("script not found: %s", scriptPath)
	}

	// Determine the shell to use
	shell := "/bin/sh"
	if _, err := os.Stat("/bin/bash"); err == nil {
		shell = "/bin/bash"
	}

	// Create the command
	cmd := exec.Command(shell, "-c", scriptPath)

	// Set working directory if provided
	if workDir != "" {
		cmd.Dir = workDir
	}

	// Set up environment
	cmd.Env = os.Environ()

	// Set timeout if specified using context
	if r.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
		defer cancel()
		// Use the context's done channel to implement timeout
		go func() {
			<-ctx.Done()
			if ctx.Err() == context.DeadlineExceeded {
				cmd.Process.Kill()
			}
		}()
	}

	// Execute and capture output
	out, err := cmd.CombinedOutput()
	stdout = strings.TrimSpace(string(out))

	// Get exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Command execution error (not just non-zero exit)
			return stdout, "", -1, fmt.Errorf("failed to execute script: %w", err)
		}
	}

	return stdout, "", exitCode, nil
}
