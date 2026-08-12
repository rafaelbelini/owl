package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents a configuration loaded from owl.config files
type Config struct {
	values map[string]string
	path   string
	exec   ExecutionConfig
}

// ExecutionConfig holds execution configuration from owl.config
// (scripts and env file path)
type ExecutionConfig struct {
	BeforeScript string
	AfterScript  string
	EnvFile      string
}

// Loader handles loading and merging configuration files
type Loader struct {
	configs map[string]*Config // keyed by directory path
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		configs: make(map[string]*Config),
	}
}

// LoadConfigForPath loads configuration files from the path and its parent directories
// More specific (deeper) configs take precedence over less specific ones
func (l *Loader) LoadConfigForPath(testPath string) error {
	// Get the directory of the test file
	testDir := filepath.Dir(testPath)

	// Find all owl.config files from test directory up to root
	dirs := l.getConfigDirs(testDir)

	// Collect execution configs and env file paths from each level
	// More specific (deeper) execution config overrides parent
	var mergedExec ExecutionConfig
	// Map from config directory to resolved env file path
	envFilesByDir := make(map[string]string) // config dir -> resolved env path

	for _, dir := range dirs {
		configPath := filepath.Join(dir, "owl.config")

		if _, err := os.Stat(configPath); err == nil {
			cfg, err := loadConfigFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config from %s: %w", configPath, err)
			}

			// Store the config for this directory
			l.configs[dir] = cfg

			// Collect execution config - more specific overrides parent
			if cfg.exec.BeforeScript != "" {
				mergedExec.BeforeScript = cfg.exec.BeforeScript
			}
			if cfg.exec.AfterScript != "" {
				mergedExec.AfterScript = cfg.exec.AfterScript
			}
			if cfg.exec.EnvFile != "" {
				mergedExec.EnvFile = cfg.exec.EnvFile
				// Resolve env path relative to the config directory
				envPath := cfg.exec.EnvFile
				if !filepath.IsAbs(envPath) {
					envPath = filepath.Join(dir, envPath)
				}
				envFilesByDir[dir] = envPath
			}
		}
	}

	// Load env files and merge values
	// More specific (deeper) env values override parent
	mergedValues := make(map[string]string)

	// Iterate from most specific to least specific (dirs is already in this order)
	for _, dir := range dirs {
		envPath, ok := envFilesByDir[dir]
		if !ok {
			continue
		}

		envValues, err := loadEnvFile(envPath)
		if err != nil {
			return fmt.Errorf("failed to load env file %s: %w", envPath, err)
		}

		// Merge values (current overrides existing)
		for k, v := range envValues {
			mergedValues[k] = v
		}
	}

	// Store the final merged config for this test path
	l.configs[testPath] = &Config{
		values: mergedValues,
		path:   testDir,
		exec:   mergedExec,
	}

	return nil
}

// getConfigDirs returns all directories from the test directory up to the root
func (l *Loader) getConfigDirs(testDir string) []string {
	var dirs []string

	// Normalize the path
	testDir = filepath.Clean(testDir)

	currentDir := testDir
	for {
		dirs = append(dirs, currentDir)

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached root
			break
		}
		currentDir = parent
	}

	// Reverse to get from root to most specific
	// Actually we want most specific first, so let's just use as-is
	// The loading logic will handle the order

	return dirs
}

// LoadConfigsForDirectory loads all owl.config files recursively from a directory
// Returns a map of directory -> Config for scoped lookups
func (l *Loader) LoadConfigsForDirectory(rootDir string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			configPath := filepath.Join(path, "owl.config")
			if _, err := os.Stat(configPath); err == nil {
				cfg, err := loadConfigFile(configPath)
				if err != nil {
					return fmt.Errorf("failed to load config from %s: %w", configPath, err)
				}
				l.configs[path] = cfg
			}
		}

		return nil
	})
}

// GetConfigForTest returns the merged configuration for a specific test path
// It considers all owl.config files from the test's directory up to root
func (l *Loader) GetConfigForTest(testPath string) (*Config, error) {
	// Check if we already loaded this test path
	if cfg, ok := l.configs[testPath]; ok {
		return cfg, nil
	}

	// Try to load now
	if err := l.LoadConfigForPath(testPath); err != nil {
		return nil, err
	}

	return l.configs[testPath], nil
}

// ResolveValues replaces placeholders in the input string with config values
// Placeholders are in the format ${key_name}
func (c *Config) ResolveValues(input string) (string, error) {
	return ResolveStringWithConfig(input, c.values)
}

// ResolveValuesInMap resolves placeholders in map values
func (c *Config) ResolveValuesInMap(m map[string]string) error {
	for key, value := range m {
		resolved, err := ResolveStringWithConfig(value, c.values)
		if err != nil {
			return fmt.Errorf("error resolving placeholder in %s: %w", key, err)
		}
		m[key] = resolved
	}
	return nil
}

// ResolveStringWithConfig resolves all placeholders in a string using the given config values
func ResolveStringWithConfig(input string, values map[string]string) (string, error) {
	result := input

	for {
		startIdx := strings.Index(result, "${")
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(result[startIdx+2:], "}")
		if endIdx == -1 {
			return "", fmt.Errorf("unclosed placeholder in: %s", input)
		}

		endIdx += startIdx + 2 // Adjust to absolute position

		placeholder := result[startIdx+2 : endIdx]
		value, ok := values[placeholder]
		if !ok {
			return "", fmt.Errorf("placeholder '${%s}' not found in configuration", placeholder)
		}

		result = result[:startIdx] + value + result[endIdx+1:]
	}

	return result, nil
}

// loadConfigFile loads a single owl.config file
func loadConfigFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	var exec ExecutionConfig
	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid config line %d: expected 'key=value', got '%s'", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("invalid config line %d: empty key", lineNum)
		}

		// Extract execution config keys
		switch key {
		case "before_all_script":
			exec.BeforeScript = value
		case "after_all_script":
			exec.AfterScript = value
		case "env":
			exec.EnvFile = value
		default:
			// All other keys go to values (for placeholder resolution)
			values[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return &Config{
		values: values,
		path:   filepath.Dir(path),
		exec:   exec,
	}, nil
}

// loadEnvFile loads variables from an env file (e.g., .env)
// Format: key=value (lines starting with # are comments, empty lines ignored)
func loadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid env line %d: expected 'key=value', got '%s' in %s", lineNum, line, path)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("invalid env line %d: empty key in %s", lineNum, path)
		}

		values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading env file: %w", err)
	}

	return values, nil
}

// HasKey checks if a key exists in the config
func (c *Config) HasKey(key string) bool {
	_, ok := c.values[key]
	return ok
}

// Get returns a value by key
func (c *Config) Get(key string) (string, bool) {
	v, ok := c.values[key]
	return v, ok
}

// GetPath returns the directory path of the config file
func (c *Config) GetPath() string {
	return c.path
}

// GetBeforeScript returns the before_all_script path from config
func (c *Config) GetBeforeScript() string {
	return c.exec.BeforeScript
}

// GetAfterScript returns the after_all_script path from config
func (c *Config) GetAfterScript() string {
	return c.exec.AfterScript
}

// GetEnvFile returns the env file path from config
func (c *Config) GetEnvFile() string {
	return c.exec.EnvFile
}
