package version

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIntegrationWithActualConfig tests the resolver with the actual jp_1.1.0.3.yaml config file
func TestIntegrationWithActualConfig(t *testing.T) {
	// Get the working directory and construct absolute path
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate up to project root and find the config file
	projectRoot := filepath.Join(wd, "..", "..")
	configPath := filepath.Join(projectRoot, "configs", "versions", "jp_1.1.0.3.yaml")

	// Check if the actual config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skipf("jp_1.1.0.3.yaml config file not found at %s, skipping integration test", configPath)
	}

	t.Run("create resolver with actual config file", func(t *testing.T) {
		resolver, err := NewConfigAddressResolver(configPath)
		if err != nil {
			t.Fatalf("NewConfigAddressResolver failed with actual config: %v", err)
		}

		if resolver == nil {
			t.Fatal("Expected resolver to be non-nil")
		}

		// Test version name
		versionName := resolver.GetVersionName()
		if versionName != "jp_1.1.0.3" {
			t.Errorf("Expected version name 'jp_1.1.0.3', got '%s'", versionName)
		}

		// Test process name (now YSO_WIN.exe)
		processName := resolver.GetProcessName()
		if processName != "YSO_WIN.exe" {
			t.Errorf("Expected process name 'YSO_WIN.exe', got '%s'", processName)
		}
	})

	t.Run("resolve actual addresses from config", func(t *testing.T) {
		resolver, err := NewConfigAddressResolver(configPath)
		if err != nil {
			t.Fatalf("NewConfigAddressResolver failed: %v", err)
		}

		// Test character address resolution
		addr, err := resolver.ResolveAddress("character")
		if err != nil {
			t.Fatalf("ResolveAddress failed for character: %v", err)
		}

		if addr == nil {
			t.Fatal("Expected address config to be non-nil for character")
		}

		// Base is now a string like "YSO_WIN.exe+2F8314"
		if baseStr, ok := addr.Base.(string); ok {
			if baseStr != "YSO_WIN.exe+2F8314" {
				t.Errorf("Expected character base 'YSO_WIN.exe+2F8314', got '%s'", baseStr)
			}
		} else {
			t.Errorf("Expected base to be string, got type %T, value: %v", addr.Base, addr.Base)
		}

		if addr.Type != "byte" {
			t.Errorf("Expected character type 'byte', got '%s'", addr.Type)
		}

		// Test hp address resolution
		addr, err = resolver.ResolveAddress("hp")
		if err != nil {
			t.Fatalf("ResolveAddress failed for hp: %v", err)
		}

		// Base is now a string like "YSO_WIN.exe+2F6ED0"
		if baseStr, ok := addr.Base.(string); ok {
			if baseStr != "YSO_WIN.exe+2F6ED0" {
				t.Errorf("Expected hp base 'YSO_WIN.exe+2F6ED0', got '%s'", baseStr)
			}
		} else {
			t.Errorf("Expected base to be string, got type %T, value: %v", addr.Base, addr.Base)
		}

		// Test player coordinates (should have base addresses)
		addr, err = resolver.ResolveAddress("player_x")
		if err != nil {
			t.Fatalf("ResolveAddress failed for player_x: %v", err)
		}

		// Base could be either numeric or string format
		if baseInt, ok := addr.Base.(int); ok {
			if baseInt != 0x00520200 {
				t.Errorf("Expected player_x base 0x00520200, got 0x%x", baseInt)
			}
		} else if baseUint64, ok := addr.Base.(uint64); ok {
			if baseUint64 != 0x00520200 {
				t.Errorf("Expected player_x base 0x00520200, got 0x%x", baseUint64)
			}
		} else if baseFloat, ok := addr.Base.(float64); ok {
			if int(baseFloat) != 0x00520200 {
				t.Errorf("Expected player_x base 0x00520200, got 0x%x", int(baseFloat))
			}
		} else {
			// Allow string format too
			t.Logf("player_x base is in string format: %v", addr.Base)
		}

		if addr.Type != "float32" {
			t.Errorf("Expected player_x type 'float32', got '%s'", addr.Type)
		}
	})

	t.Run("test NewConfigAddressResolverByVersion with cwd", func(t *testing.T) {
		// Change to project root temporarily
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)

		projectRoot := filepath.Join(originalWd, "..", "..")
		if err := os.Chdir(projectRoot); err != nil {
			t.Fatalf("Failed to change directory to project root: %v", err)
		}

		resolver, err := NewConfigAddressResolverByVersion("jp_1.1.0.3")
		if err != nil {
			t.Fatalf("NewConfigAddressResolverByVersion failed: %v", err)
		}

		// Verify basic functionality
		versionName := resolver.GetVersionName()
		if versionName != "jp_1.1.0.3" {
			t.Errorf("Expected version name 'jp_1.1.0.3', got '%s'", versionName)
		}

		// Verify address resolution works
		_, err = resolver.ResolveAddress("character")
		if err != nil {
			t.Fatalf("ResolveAddress failed for character: %v", err)
		}
	})
}
