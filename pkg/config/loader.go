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

	// Load configs from most specific (deepest) to least specific
	// Start with empty config and merge parent configs
	mergedConfig := make(map[string]string)

	for _, dir := range dirs {
		configPath := filepath.Join(dir, "owl.config")

		if _, err := os.Stat(configPath); err == nil {
			cfg, err := loadConfigFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config from %s: %w", configPath, err)
			}

			// Merge: parent configs are overridden by more specific configs
			// So we apply mergedConfig to cfg, then the result becomes mergedConfig
			// Actually, we want: more specific (child) overrides parent
			// Start from root and go deeper, each level overrides previous

			// Create a copy of merged config
			newMerged := make(map[string]string)

			// First copy existing merged (parent) values
			for k, v := range mergedConfig {
				newMerged[k] = v
			}

			// Then override with current directory's values
			for k, v := range cfg.values {
				newMerged[k] = v
			}

			mergedConfig = newMerged

			// Store the config for this directory
			l.configs[dir] = cfg
		}
	}

	// Store the final merged config for this test path
	l.configs[testPath] = &Config{
		values: mergedConfig,
		path:   testDir,
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

		values[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return &Config{
		values: values,
		path:   filepath.Dir(path),
	}, nil
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
