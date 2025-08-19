package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadVersionConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	
	t.Run("successful load", func(t *testing.T) {
		// Create a valid test config file
		configPath := filepath.Join(tempDir, "test_version.yaml")
		configContent := `
metadata:
  name: test_version
  display_name: Test Version
  description: Test description
  process_name: test.exe

addresses:
  test_address:
    base: 0x12345678
    offsets: [0x10, 0x20]
    type: int32
    description: Test address
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		config, err := LoadVersionConfig(configPath)
		if err != nil {
			t.Fatalf("LoadVersionConfig failed: %v", err)
		}
		
		// Verify metadata
		if config.Metadata.Name != "test_version" {
			t.Errorf("Expected name 'test_version', got '%s'", config.Metadata.Name)
		}
		if config.Metadata.ProcessName != "test.exe" {
			t.Errorf("Expected process_name 'test.exe', got '%s'", config.Metadata.ProcessName)
		}
		
		// Verify addresses
		if len(config.Addresses) != 1 {
			t.Errorf("Expected 1 address, got %d", len(config.Addresses))
		}
		if addr, ok := config.Addresses["test_address"]; ok {
			// Base is now interface{}, need to handle type assertion
			if baseInt, ok := addr.Base.(int); ok {
				if baseInt != 0x12345678 {
					t.Errorf("Expected base 0x12345678, got 0x%x", baseInt)
				}
			} else if baseUint64, ok := addr.Base.(uint64); ok {
				// YAML might parse as uint64
				if baseUint64 != 0x12345678 {
					t.Errorf("Expected base 0x12345678, got 0x%x", baseUint64)
				}
			} else if baseFloat, ok := addr.Base.(float64); ok {
				// YAML might parse as float64
				if int(baseFloat) != 0x12345678 {
					t.Errorf("Expected base 0x12345678, got 0x%x", int(baseFloat))
				}
			} else {
				t.Errorf("Unexpected base type: %T", addr.Base)
			}
			if len(addr.Offsets) != 2 {
				t.Errorf("Expected 2 offsets, got %d", len(addr.Offsets))
			}
			if addr.Type != "int32" {
				t.Errorf("Expected type 'int32', got '%s'", addr.Type)
			}
		} else {
			t.Error("Address 'test_address' not found")
		}
	})
	
	t.Run("empty path", func(t *testing.T) {
		_, err := LoadVersionConfig("")
		if err == nil {
			t.Error("Expected error for empty path, got nil")
		}
	})
	
	t.Run("non-existent file", func(t *testing.T) {
		_, err := LoadVersionConfig(filepath.Join(tempDir, "non_existent.yaml"))
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
	
	t.Run("invalid YAML", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "invalid.yaml")
		configContent := `
this is not: valid yaml
  - because of: inconsistent
    indentation and structure
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		_, err := LoadVersionConfig(configPath)
		if err == nil {
			t.Error("Expected error for invalid YAML, got nil")
		}
	})
	
	t.Run("missing required metadata.name", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "missing_name.yaml")
		configContent := `
metadata:
  display_name: Test Version
  process_name: test.exe
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		_, err := LoadVersionConfig(configPath)
		if err == nil {
			t.Error("Expected error for missing metadata.name, got nil")
		}
	})
	
	t.Run("missing required metadata.process_name", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "missing_process.yaml")
		configContent := `
metadata:
  name: test_version
  display_name: Test Version
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		_, err := LoadVersionConfig(configPath)
		if err == nil {
			t.Error("Expected error for missing metadata.process_name, got nil")
		}
	})
}

func TestLoadAppConfig(t *testing.T) {
	tempDir := t.TempDir()
	
	t.Run("default config with empty path", func(t *testing.T) {
		config, err := LoadAppConfig("")
		if err != nil {
			t.Fatalf("LoadAppConfig failed: %v", err)
		}
		
		if config.Version != "jp_1.1.0.3" {
			t.Errorf("Expected default version 'jp_1.1.0.3', got '%s'", config.Version)
		}
		if config.Port != 8080 {
			t.Errorf("Expected default port 8080, got %d", config.Port)
		}
		if config.Interval != 100*time.Millisecond {
			t.Errorf("Expected default interval 100ms, got %v", config.Interval)
		}
	})
	
	t.Run("successful load from file", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "app.yaml")
		configContent := `
version: custom_version
port: 9090
interval: "200ms"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		config, err := LoadAppConfig(configPath)
		if err != nil {
			t.Fatalf("LoadAppConfig failed: %v", err)
		}
		
		if config.Version != "custom_version" {
			t.Errorf("Expected version 'custom_version', got '%s'", config.Version)
		}
		if config.Port != 9090 {
			t.Errorf("Expected port 9090, got %d", config.Port)
		}
		if config.Interval != 200*time.Millisecond {
			t.Errorf("Expected interval 200ms, got %v", config.Interval)
		}
	})
	
	t.Run("non-existent file", func(t *testing.T) {
		_, err := LoadAppConfig(filepath.Join(tempDir, "non_existent.yaml"))
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
	
	t.Run("invalid port", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "invalid_port.yaml")
		configContent := `
version: test
port: 70000
interval: 100000000
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		_, err := LoadAppConfig(configPath)
		if err == nil {
			t.Error("Expected error for invalid port, got nil")
		}
	})
	
	t.Run("interval too short", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "short_interval.yaml")
		configContent := `
version: test
port: 8080
interval: "5ms"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		_, err := LoadAppConfig(configPath)
		if err == nil {
			t.Error("Expected error for interval too short, got nil")
		}
	})
}

func TestLoadVersionConfigByName(t *testing.T) {
	// This test requires the actual configs/versions directory structure
	// We'll test with the jp_1.1.0.3.yaml file if it exists
	
	t.Run("load existing version config", func(t *testing.T) {
		// Check if the config file exists
		configPath := filepath.Join("configs", "versions", "jp_1.1.0.3.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Create the directories and file for testing
			if err := os.MkdirAll(filepath.Join("configs", "versions"), 0755); err != nil {
				t.Skipf("Cannot create test directories: %v", err)
			}
			
			// Use minimal config for testing
			configContent := `
metadata:
  name: jp_1.1.0.3
  display_name: Japanese v1.1.0.3
  process_name: yso_win.exe

addresses:
  test:
    base: 0x12345678
    offsets: []
    type: int32
    description: Test address
`
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Skipf("Cannot create test file: %v", err)
			}
			defer os.Remove(configPath)
		}
		
		config, err := LoadVersionConfigByName("jp_1.1.0.3")
		if err != nil {
			t.Fatalf("LoadVersionConfigByName failed: %v", err)
		}
		
		if config.Metadata.Name != "jp_1.1.0.3" {
			t.Errorf("Expected name 'jp_1.1.0.3', got '%s'", config.Metadata.Name)
		}
		if config.Metadata.ProcessName != "yso_win.exe" {
			t.Errorf("Expected process_name 'yso_win.exe', got '%s'", config.Metadata.ProcessName)
		}
	})
	
	t.Run("non-existent version", func(t *testing.T) {
		_, err := LoadVersionConfigByName("non_existent_version")
		if err == nil {
			t.Error("Expected error for non-existent version, got nil")
		}
	})
}