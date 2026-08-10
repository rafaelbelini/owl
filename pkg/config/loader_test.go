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
	configsDir := filepath.Join(tmpDir, ".configs")
	if err := os.Mkdir(configsDir, 0755); err != nil {
		t.Fatalf("Failed to create .configs dir: %v", err)
	}
	configPath := filepath.Join(configsDir, "owl.config")

	configContent := `# Test config
api_url=https://api.example.com
api_token=Bearer test_token
port=8080
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	cfg, err := loadConfigFile(configPath)
	if err != nil {
		t.Fatalf("loadConfigFile() error = %v", err)
	}

	if cfg.values["api_url"] != "https://api.example.com" {
		t.Errorf("api_url = %v, want https://api.example.com", cfg.values["api_url"])
	}
	if cfg.values["api_token"] != "Bearer test_token" {
		t.Errorf("api_token = %v, want Bearer test_token", cfg.values["api_token"])
	}
	if cfg.values["port"] != "8080" {
		t.Errorf("port = %v, want 8080", cfg.values["port"])
	}
}

func TestLoadConfigFileInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	configsDir := filepath.Join(tmpDir, ".configs")
	if err := os.Mkdir(configsDir, 0755); err != nil {
		t.Fatalf("Failed to create .configs dir: %v", err)
	}
	configPath := filepath.Join(configsDir, "owl.config")

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

	// Create parent directory with config
	parentConfigsDir := filepath.Join(tmpDir, ".configs")
	if err := os.Mkdir(parentConfigsDir, 0755); err != nil {
		t.Fatalf("Failed to create parent .configs dir: %v", err)
	}
	parentConfig := `api_url=https://parent.example.com`
	parentPath := filepath.Join(parentConfigsDir, "owl.config")
	if err := os.WriteFile(parentPath, []byte(parentConfig), 0644); err != nil {
		t.Fatalf("Failed to create parent config: %v", err)
	}

	// Create child directory with config
	childDir := filepath.Join(tmpDir, "child")
	if err := os.Mkdir(childDir, 0755); err != nil {
		t.Fatalf("Failed to create child dir: %v", err)
	}

	childConfigsDir := filepath.Join(childDir, ".configs")
	if err := os.Mkdir(childConfigsDir, 0755); err != nil {
		t.Fatalf("Failed to create child .configs dir: %v", err)
	}

	childConfig := `api_token=Bearer child_token`
	childPath := filepath.Join(childConfigsDir, "owl.config")
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

	// Should have both parent and child values
	if cfg.values["api_url"] != "https://parent.example.com" {
		t.Errorf("api_url = %v, want https://parent.example.com", cfg.values["api_url"])
	}
	if cfg.values["api_token"] != "Bearer child_token" {
		t.Errorf("api_token = %v, want Bearer child_token", cfg.values["api_token"])
	}
}
