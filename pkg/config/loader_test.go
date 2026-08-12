package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveStringWithConfig(t *testing.T) {
	values := map[string]string{
		"api_url":  "https://api.example.com",
		"api_token": "Bearer token123",
		"port":     "8080",
	}

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple placeholder",
			input:    "https://${api_url}/users",
			expected: "https://https://api.example.com/users",
			wantErr:  false,
		},
		{
			name:     "multiple placeholders",
			input:    "${api_url}:${port}/api",
			expected: "https://api.example.com:8080/api",
			wantErr:  false,
		},
		{
			name:     "placeholder at start",
			input:    "${api_token} header",
			expected: "Bearer token123 header",
			wantErr:  false,
		},
		{
			name:     "no placeholder",
			input:    "https://api.example.com/users",
			expected: "https://api.example.com/users",
			wantErr:  false,
		},
		{
			name:     "missing placeholder",
			input:    "${nonexistent}",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "unclosed placeholder",
			input:    "${api_url",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveStringWithConfig(tt.input, values)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveStringWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ResolveStringWithConfig() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoadConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".owl")

	// .owl now only contains execution config (before_all_script, after_all_script, env)
	// Variables go in the env file
	configContent := `# Test config
before_all_script=./setup.sh
after_all_script=./teardown.sh
env=.env
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	cfg, err := loadConfigFile(configPath)
	if err != nil {
		t.Fatalf("loadConfigFile() error = %v", err)
	}

	// Execution config should be set
	if cfg.exec.BeforeScript != "./setup.sh" {
		t.Errorf("BeforeScript = %v, want ./setup.sh", cfg.exec.BeforeScript)
	}
	if cfg.exec.AfterScript != "./teardown.sh" {
		t.Errorf("AfterScript = %v, want ./teardown.sh", cfg.exec.AfterScript)
	}
	if cfg.exec.EnvFile != ".env" {
		t.Errorf("EnvFile = %v, want .env", cfg.exec.EnvFile)
	}

	// No values should be in .owl anymore (they go in env file)
	if len(cfg.values) != 0 {
		t.Errorf("values should be empty, got %v", cfg.values)
	}
}

func TestLoadConfigFileInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".owl")

	// Invalid line (no =)
	invalidContent := `api_url`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	_, err := loadConfigFile(configPath)
	if err == nil {
		t.Error("loadConfigFile() expected error for invalid line")
	}
}

func TestLoadConfigForPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create child directory with .owl
	childDir := filepath.Join(tmpDir, "child")
	if err := os.Mkdir(childDir, 0755); err != nil {
		t.Fatalf("Failed to create child dir: %v", err)
	}

	// Child has .owl with env pointing to .env
	childConfig := `env=.env`
	childPath := filepath.Join(childDir, ".owl")
	if err := os.WriteFile(childPath, []byte(childConfig), 0644); err != nil {
		t.Fatalf("Failed to create child config: %v", err)
	}

	// Child .env file
	childEnv := `api_token=Bearer child_token`
	childEnvPath := filepath.Join(childDir, ".env")
	if err := os.WriteFile(childEnvPath, []byte(childEnv), 0644); err != nil {
		t.Fatalf("Failed to create child .env: %v", err)
	}

	// Create test file in child directory
	testFile := filepath.Join(childDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	if err := loader.LoadConfigForPath(testFile); err != nil {
		t.Fatalf("LoadConfigForPath() error = %v", err)
	}

	cfg, err := loader.GetConfigForTest(testFile)
	if err != nil {
		t.Fatalf("GetConfigForTest() error = %v", err)
	}

	// Should have child values (from explicit env config)
	if cfg.values["api_token"] != "Bearer child_token" {
		t.Errorf("api_token = %v, want Bearer child_token", cfg.values["api_token"])
	}
}

func TestLoadConfigForPathDefaultEnv(t *testing.T) {
	tmpDir := t.TempDir()

	// Create child directory with .owl (no env config)
	childDir := filepath.Join(tmpDir, "child")
	if err := os.Mkdir(childDir, 0755); err != nil {
		t.Fatalf("Failed to create child dir: %v", err)
	}

	// Child has .owl WITHOUT env config
	childConfig := `before_all_script=./setup.sh`
	childPath := filepath.Join(childDir, ".owl")
	if err := os.WriteFile(childPath, []byte(childConfig), 0644); err != nil {
		t.Fatalf("Failed to create child config: %v", err)
	}

	// Child .env file (should be auto-discovered)
	childEnv := `api_url=https://child.example.com`
	childEnvPath := filepath.Join(childDir, ".env")
	if err := os.WriteFile(childEnvPath, []byte(childEnv), 0644); err != nil {
		t.Fatalf("Failed to create child .env: %v", err)
	}

	// Create test file in child directory
	testFile := filepath.Join(childDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	if err := loader.LoadConfigForPath(testFile); err != nil {
		t.Fatalf("LoadConfigForPath() error = %v", err)
	}

	cfg, err := loader.GetConfigForTest(testFile)
	if err != nil {
		t.Fatalf("GetConfigForTest() error = %v", err)
	}

	// Should have child values (from default .env discovery)
	if cfg.values["api_url"] != "https://child.example.com" {
		t.Errorf("api_url = %v, want https://child.example.com", cfg.values["api_url"])
	}
}

func TestLoadConfigForPathNoEnvFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create child directory with .owl (no env config, no .env file)
	childDir := filepath.Join(tmpDir, "child")
	if err := os.Mkdir(childDir, 0755); err != nil {
		t.Fatalf("Failed to create child dir: %v", err)
	}

	// Child has .owl WITHOUT env config
	childConfig := `before_all_script=./setup.sh`
	childPath := filepath.Join(childDir, ".owl")
	if err := os.WriteFile(childPath, []byte(childConfig), 0644); err != nil {
		t.Fatalf("Failed to create child config: %v", err)
	}

	// Create test file in child directory
	testFile := filepath.Join(childDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	if err := loader.LoadConfigForPath(testFile); err != nil {
		t.Fatalf("LoadConfigForPath() error = %v", err)
	}

	cfg, err := loader.GetConfigForTest(testFile)
	if err != nil {
		t.Fatalf("GetConfigForTest() error = %v", err)
	}

	// Should have no values (no env file)
	if len(cfg.values) != 0 {
		t.Errorf("values should be empty, got %v", cfg.values)
	}

	// But should still have script config
	if cfg.exec.BeforeScript != "./setup.sh" {
		t.Errorf("BeforeScript = %v, want ./setup.sh", cfg.exec.BeforeScript)
	}
}

func TestLoadEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envContent := `# Test env file
api_url=https://api.example.com
api_token=Bearer test_token
port=8080
`
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("Failed to create temp env file: %v", err)
	}

	values, err := loadEnvFile(envPath)
	if err != nil {
		t.Fatalf("loadEnvFile() error = %v", err)
	}

	if values["api_url"] != "https://api.example.com" {
		t.Errorf("api_url = %v, want https://api.example.com", values["api_url"])
	}
	if values["api_token"] != "Bearer test_token" {
		t.Errorf("api_token = %v, want Bearer test_token", values["api_token"])
	}
	if values["port"] != "8080" {
		t.Errorf("port = %v, want 8080", values["port"])
	}
}

func TestLoadEnvFileInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	// Invalid line (no =)
	invalidContent := `api_url`
	if err := os.WriteFile(envPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create temp env file: %v", err)
	}

	_, err := loadEnvFile(envPath)
	if err == nil {
		t.Error("loadEnvFile() expected error for invalid line")
	}
}
