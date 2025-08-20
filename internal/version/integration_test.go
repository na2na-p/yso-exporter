package version

import (
	"testing"
)

// TestIntegrationWithBuiltinConfig tests the resolver with built-in configurations
func TestIntegrationWithBuiltinConfig(t *testing.T) {
	t.Run("create resolver with built-in jp_1.1.0.3 config", func(t *testing.T) {
		resolver, err := NewConfigAddressResolverByVersion("jp_1.1.0.3")
		if err != nil {
			t.Fatalf("NewConfigAddressResolverByVersion failed: %v", err)
		}

		if resolver == nil {
			t.Fatal("Expected resolver to be non-nil")
		}

		// Test version name
		versionName := resolver.GetVersionName()
		if versionName != "jp_1.1.0.3" {
			t.Errorf("Expected version name 'jp_1.1.0.3', got '%s'", versionName)
		}

		// Test process name
		processName := resolver.GetProcessName()
		if processName != "YSO_WIN.exe" {
			t.Errorf("Expected process name 'YSO_WIN.exe', got '%s'", processName)
		}
	})

	t.Run("resolve actual addresses from built-in config", func(t *testing.T) {
		resolver, err := NewConfigAddressResolverByVersion("jp_1.1.0.3")
		if err != nil {
			t.Fatalf("Failed to create resolver: %v", err)
		}

		// Test character address
		charAddr, err := resolver.ResolveAddress("character")
		if err != nil {
			t.Errorf("Failed to resolve character address: %v", err)
		} else {
			if charAddr.Base != "YSO_WIN.exe+2F8314" {
				t.Errorf("Expected character base 'YSO_WIN.exe+2F8314', got '%s'", charAddr.Base)
			}
			if charAddr.Type != "byte" {
				t.Errorf("Expected character type 'byte', got '%s'", charAddr.Type)
			}
		}

		// Test hp address
		hpAddr, err := resolver.ResolveAddress("hp")
		if err != nil {
			t.Errorf("Failed to resolve hp address: %v", err)
		} else {
			if hpAddr.Base != "YSO_WIN.exe+2F6ED0" {
				t.Errorf("Expected hp base 'YSO_WIN.exe+2F6ED0', got '%s'", hpAddr.Base)
			}
			if hpAddr.Type != "float32" {
				t.Errorf("Expected hp type 'float32', got '%s'", hpAddr.Type)
			}
		}

		// Test mp address
		mpAddr, err := resolver.ResolveAddress("mp")
		if err != nil {
			t.Errorf("Failed to resolve mp address: %v", err)
		} else {
			if mpAddr.Base != "YSO_WIN.exe+2F6ED8" {
				t.Errorf("Expected mp base 'YSO_WIN.exe+2F6ED8', got '%s'", mpAddr.Base)
			}
			if mpAddr.Type != "float32" {
				t.Errorf("Expected mp type 'float32', got '%s'", mpAddr.Type)
			}
		}

		// Test level address
		levelAddr, err := resolver.ResolveAddress("level")
		if err != nil {
			t.Errorf("Failed to resolve level address: %v", err)
		} else {
			if levelAddr.Base != "YSO_WIN.exe+2F6F00" {
				t.Errorf("Expected level base 'YSO_WIN.exe+2F6F00', got '%s'", levelAddr.Base)
			}
			if levelAddr.Type != "int32" {
				t.Errorf("Expected level type 'int32', got '%s'", levelAddr.Type)
			}
		}

		// Test experience address
		expAddr, err := resolver.ResolveAddress("experience")
		if err != nil {
			t.Errorf("Failed to resolve experience address: %v", err)
		} else {
			if expAddr.Base != "YSO_WIN.exe+2F6EE8" {
				t.Errorf("Expected experience base 'YSO_WIN.exe+2F6EE8', got '%s'", expAddr.Base)
			}
			if expAddr.Type != "float32" {
				t.Errorf("Expected experience type 'float32', got '%s'", expAddr.Type)
			}
		}

		// Test non-existent address
		_, err = resolver.ResolveAddress("non_existent_metric")
		if err == nil {
			t.Error("Expected error when resolving non-existent metric")
		}
	})

	t.Run("create resolver with built-in jp_1.0.0.4 config", func(t *testing.T) {
		resolver, err := NewConfigAddressResolverByVersion("jp_1.0.0.4")
		if err != nil {
			t.Fatalf("NewConfigAddressResolverByVersion failed: %v", err)
		}

		if resolver == nil {
			t.Fatal("Expected resolver to be non-nil")
		}

		// Test version name
		versionName := resolver.GetVersionName()
		if versionName != "jp_1.0.0.4" {
			t.Errorf("Expected version name 'jp_1.0.0.4', got '%s'", versionName)
		}

		// Test process name
		processName := resolver.GetProcessName()
		if processName != "YSO_WIN.exe" {
			t.Errorf("Expected process name 'YSO_WIN.exe', got '%s'", processName)
		}

		// Test character address (different from jp_1.1.0.3)
		charAddr, err := resolver.ResolveAddress("character")
		if err != nil {
			t.Errorf("Failed to resolve character address: %v", err)
		} else {
			if charAddr.Base != "YSO_WIN.exe+2E241C" {
				t.Errorf("Expected character base 'YSO_WIN.exe+2E241C', got '%s'", charAddr.Base)
			}
		}
	})
}
