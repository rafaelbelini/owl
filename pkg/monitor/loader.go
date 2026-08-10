package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rafael/owl/cli/pkg/config"
	"gopkg.in/yaml.v3"
)

// LoadTest loads a test configuration from a YAML file
func LoadTest(path string) (TestConfig, error) {
	return LoadTestWithConfig(path, nil)
}

// LoadTestWithConfig loads a test configuration from a YAML file and resolves placeholders using config
func LoadTestWithConfig(path string, cfg *config.Config) (TestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TestConfig{}, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// Convert to string for config resolution
	content := string(data)

	// Resolve placeholders if config is provided
	if cfg != nil {
		content, err = cfg.ResolveValues(content)
		if err != nil {
			return TestConfig{}, fmt.Errorf("failed to resolve config in %s: %w", path, err)
		}
	}

	var config TestConfig
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return TestConfig{}, fmt.Errorf("failed to parse YAML in %s: %w", path, err)
	}

	// Set defaults
	if config.Request.Method == "" {
		config.Request.Method = "GET"
	}
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 10
	}

	// Set the working directory to the test file's directory
	config.WorkDir = filepath.Dir(path)

	return config, nil
}

// FindTestFiles finds all .yaml and .yml files in a path (file or directory)
func FindTestFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %s", path)
	}

	var files []string

	if info.IsDir() {
		// Walk directory recursively
		err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && isHTTPTestFile(p) {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking directory: %w", err)
		}
	} else {
		// Single file
		if isHTTPTestFile(path) {
			files = append(files, path)
		} else {
			return nil, fmt.Errorf("file must be a HTTP test YAML file (must contain 'request:'): %s", path)
		}
	}

	return files, nil
}

// isTestFile checks if a file has a valid test extension
func isTestFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

// isHTTPTestFile checks if a file is an HTTP test by looking for "request:" field
func isHTTPTestFile(filePath string) bool {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	content := strings.ToLower(string(data))
	return strings.Contains(content, "request:")
}
