package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfigAddressResolver(t *testing.T) {
	tempDir := t.TempDir()
	
	t.Run("successful creation with valid config", func(t *testing.T) {
		// Create a valid test config file
		configPath := filepath.Join(tempDir, "test.yaml")
		configContent := `
metadata:
  name: test_version
  display_name: Test Version
  description: Test description
  process_name: test.exe

addresses:
  character_id:
    base: 0x0052CDE8
    offsets: []
    type: int32
    description: "Current character ID"
  
  player_x:
    base: 0x00520200
    offsets: [0x10, 0x20]
    type: float32
    description: "Player X coordinate"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		resolver, err := NewConfigAddressResolver(configPath)
		if err != nil {
			t.Fatalf("NewConfigAddressResolver failed: %v", err)
		}
		
		if resolver == nil {
			t.Fatal("Expected resolver to be non-nil")
		}
	})
	
	t.Run("creation fails with invalid config path", func(t *testing.T) {
		_, err := NewConfigAddressResolver("non_existent.yaml")
		if err == nil {
			t.Error("Expected error for non-existent config file, got nil")
		}
	})
	
	t.Run("creation fails with empty config path", func(t *testing.T) {
		_, err := NewConfigAddressResolver("")
		if err == nil {
			t.Error("Expected error for empty config path, got nil")
		}
	})
}

func TestConfigAddressResolver_ResolveAddress(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a valid test config file
	configPath := filepath.Join(tempDir, "test.yaml")
	configContent := `
metadata:
  name: test_version
  display_name: Test Version  
  description: Test description
  process_name: test.exe

addresses:
  character_id:
    base: 0x0052CDE8
    offsets: []
    type: int32
    description: "Current character ID"
  
  player_x:
    base: 0x00520200
    offsets: [0x10, 0x20]
    type: float32
    description: "Player X coordinate"
    
  nested_value:
    base: 0x00123456
    offsets: [0x00, 0x08, 0x0C]
    type: int32
    description: "Nested pointer value"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	resolver, err := NewConfigAddressResolver(configPath)
	if err != nil {
		t.Fatalf("NewConfigAddressResolver failed: %v", err)
	}
	
	t.Run("successful resolution of existing metric", func(t *testing.T) {
		addr, err := resolver.ResolveAddress("character_id")
		if err != nil {
			t.Fatalf("ResolveAddress failed: %v", err)
		}
		
		if addr == nil {
			t.Fatal("Expected address config to be non-nil")
		}
		
		// Base is now interface{}, need to handle type assertion
		if baseInt, ok := addr.Base.(int); ok {
			if baseInt != 0x0052CDE8 {
				t.Errorf("Expected base 0x0052CDE8, got 0x%x", baseInt)
			}
		} else if baseUint64, ok := addr.Base.(uint64); ok {
			// YAML might parse as uint64
			if baseUint64 != 0x0052CDE8 {
				t.Errorf("Expected base 0x0052CDE8, got 0x%x", baseUint64)
			}
		} else if baseFloat, ok := addr.Base.(float64); ok {
			// YAML might parse as float64
			if int(baseFloat) != 0x0052CDE8 {
				t.Errorf("Expected base 0x0052CDE8, got 0x%x", int(baseFloat))
			}
		} else {
			t.Errorf("Unexpected base type: %T, value: %v", addr.Base, addr.Base)
		}
		
		if len(addr.Offsets) != 0 {
			t.Errorf("Expected 0 offsets, got %d", len(addr.Offsets))
		}
		
		if addr.Type != "int32" {
			t.Errorf("Expected type 'int32', got '%s'", addr.Type)
		}
		
		if addr.Description != "Current character ID" {
			t.Errorf("Expected description 'Current character ID', got '%s'", addr.Description)
		}
	})
	
	t.Run("successful resolution with offsets", func(t *testing.T) {
		addr, err := resolver.ResolveAddress("player_x")
		if err != nil {
			t.Fatalf("ResolveAddress failed: %v", err)
		}
		
		// Base is now interface{}, need to handle type assertion
		if baseInt, ok := addr.Base.(int); ok {
			if baseInt != 0x00520200 {
				t.Errorf("Expected base 0x00520200, got 0x%x", baseInt)
			}
		} else if baseUint64, ok := addr.Base.(uint64); ok {
			// YAML might parse as uint64
			if baseUint64 != 0x00520200 {
				t.Errorf("Expected base 0x00520200, got 0x%x", baseUint64)
			}
		} else if baseFloat, ok := addr.Base.(float64); ok {
			// YAML might parse as float64
			if int(baseFloat) != 0x00520200 {
				t.Errorf("Expected base 0x00520200, got 0x%x", int(baseFloat))
			}
		} else {
			t.Errorf("Unexpected base type: %T, value: %v", addr.Base, addr.Base)
		}
		
		if len(addr.Offsets) != 2 {
			t.Errorf("Expected 2 offsets, got %d", len(addr.Offsets))
		}
		
		expectedOffsets := []uint32{0x10, 0x20}
		for i, expected := range expectedOffsets {
			if addr.Offsets[i] != expected {
				t.Errorf("Expected offset[%d] 0x%x, got 0x%x", i, expected, addr.Offsets[i])
			}
		}
		
		if addr.Type != "float32" {
			t.Errorf("Expected type 'float32', got '%s'", addr.Type)
		}
	})
	
	t.Run("successful resolution with multiple offsets", func(t *testing.T) {
		addr, err := resolver.ResolveAddress("nested_value")
		if err != nil {
			t.Fatalf("ResolveAddress failed: %v", err)
		}
		
		if len(addr.Offsets) != 3 {
			t.Errorf("Expected 3 offsets, got %d", len(addr.Offsets))
		}
		
		expectedOffsets := []uint32{0x00, 0x08, 0x0C}
		for i, expected := range expectedOffsets {
			if addr.Offsets[i] != expected {
				t.Errorf("Expected offset[%d] 0x%x, got 0x%x", i, expected, addr.Offsets[i])
			}
		}
	})
	
	t.Run("resolution fails for non-existent metric", func(t *testing.T) {
		_, err := resolver.ResolveAddress("non_existent_metric")
		if err == nil {
			t.Error("Expected error for non-existent metric, got nil")
		}
	})
	
	t.Run("resolution fails for empty metric name", func(t *testing.T) {
		_, err := resolver.ResolveAddress("")
		if err == nil {
			t.Error("Expected error for empty metric name, got nil")
		}
	})
}

func TestConfigAddressResolver_GetProcessName(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a valid test config file
	configPath := filepath.Join(tempDir, "test.yaml")
	configContent := `
metadata:
  name: test_version
  display_name: Test Version
  description: Test description
  process_name: yso_win.exe

addresses:
  test_address:
    base: 0x12345678
    offsets: []
    type: int32
    description: Test address
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	resolver, err := NewConfigAddressResolver(configPath)
	if err != nil {
		t.Fatalf("NewConfigAddressResolver failed: %v", err)
	}
	
	processName := resolver.GetProcessName()
	if processName != "yso_win.exe" {
		t.Errorf("Expected process name 'yso_win.exe', got '%s'", processName)
	}
}

func TestConfigAddressResolver_GetVersionName(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a valid test config file  
	configPath := filepath.Join(tempDir, "test.yaml")
	configContent := `
metadata:
  name: jp_1.1.0.3
  display_name: Japanese Version 1.1.0.3
  description: Test description
  process_name: yso_win.exe

addresses:
  test_address:
    base: 0x12345678
    offsets: []
    type: int32
    description: Test address
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	resolver, err := NewConfigAddressResolver(configPath)
	if err != nil {
		t.Fatalf("NewConfigAddressResolver failed: %v", err)
	}
	
	versionName := resolver.GetVersionName()
	if versionName != "jp_1.1.0.3" {
		t.Errorf("Expected version name 'jp_1.1.0.3', got '%s'", versionName)
	}
}

func TestNewConfigAddressResolverByVersion(t *testing.T) {
	t.Run("successful creation with existing version", func(t *testing.T) {
		// This test assumes the jp_1.1.0.3.yaml file exists in configs/versions/
		// Skip if the file doesn't exist
		configPath := filepath.Join("configs", "versions", "jp_1.1.0.3.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Skip("jp_1.1.0.3.yaml config file not found, skipping test")
		}
		
		resolver, err := NewConfigAddressResolverByVersion("jp_1.1.0.3")
		if err != nil {
			t.Fatalf("NewConfigAddressResolverByVersion failed: %v", err)
		}
		
		if resolver == nil {
			t.Fatal("Expected resolver to be non-nil")
		}
		
		// Verify it can resolve some addresses from the actual config
		versionName := resolver.GetVersionName()
		if versionName != "jp_1.1.0.3" {
			t.Errorf("Expected version name 'jp_1.1.0.3', got '%s'", versionName)
		}
		
		processName := resolver.GetProcessName()
		if processName != "yso_win.exe" {
			t.Errorf("Expected process name 'yso_win.exe', got '%s'", processName)
		}
		
		// Test resolving a known address from the config
		addr, err := resolver.ResolveAddress("character_id")
		if err != nil {
			t.Fatalf("ResolveAddress failed for character_id: %v", err)
		}
		
		if addr == nil {
			t.Fatal("Expected address config to be non-nil for character_id")
		}
	})
	
	t.Run("creation fails with non-existent version", func(t *testing.T) {
		_, err := NewConfigAddressResolverByVersion("non_existent_version")
		if err == nil {
			t.Error("Expected error for non-existent version, got nil")
		}
	})
	
	t.Run("creation fails with empty version name", func(t *testing.T) {
		_, err := NewConfigAddressResolverByVersion("")
		if err == nil {
			t.Error("Expected error for empty version name, got nil")
		}
	})
}