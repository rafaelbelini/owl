package script

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	// Resolve relative paths against workDir
	if !filepath.IsAbs(scriptPath) && workDir != "" {
		scriptPath = filepath.Join(workDir, scriptPath)
	}

	// Validate script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", "", -1, fmt.Errorf("script not found: %s", scriptPath)
	}

	// Determine the shell to use
	shell := "/bin/sh"
	if _, err := os.Stat("/bin/bash"); err == nil {
		shell = "/bin/bash"
	}

	// Build the command - we need to ensure the script is found in workDir
	// Problem: filepath.Join(".", "./script.sh") returns "script.sh" (loses "./")
	// bash -c "script.sh" searches PATH, not the current directory
	// Fix: prepend "./" for relative paths so bash finds the script in workDir
	cmdStr := scriptPath
	if !filepath.IsAbs(scriptPath) && workDir != "" {
		// Ensure script has a path prefix so bash finds it in workDir, not PATH
		if !strings.HasPrefix(filepath.Base(scriptPath), ".") {
			cmdStr = "./" + scriptPath
		}
		// Use cd to workDir so relative paths are resolved correctly
		cmdStr = "cd " + workDir + " && " + cmdStr
	}
	cmd := exec.Command(shell, "-c", cmdStr)

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
