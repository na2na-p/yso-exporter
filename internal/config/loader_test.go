package config_test

import (
	"testing"

	"github.com/na2na/yso-exporter/internal/config"
)

func TestLoadVersionConfigByName(t *testing.T) {
	t.Run("successful load - jp_1.1.0.3", func(t *testing.T) {
		cfg, err := config.LoadVersionConfigByName("jp_1.1.0.3")
		if err != nil {
			t.Fatalf("LoadVersionConfigByName failed: %v", err)
		}

		// Verify metadata
		if cfg.Metadata.Name != "jp_1.1.0.3" {
			t.Errorf("Expected name 'jp_1.1.0.3', got '%s'", cfg.Metadata.Name)
		}
		if cfg.Metadata.ProcessName != "YSO_WIN.exe" {
			t.Errorf("Expected process_name 'YSO_WIN.exe', got '%s'", cfg.Metadata.ProcessName)
		}

		// Verify some addresses exist
		if len(cfg.Addresses) == 0 {
			t.Errorf("Expected addresses to be loaded")
		}

		// Check specific address
		if addr, ok := cfg.Addresses["character"]; !ok {
			t.Errorf("Expected 'character' address to exist")
		} else {
			if addr.Type != "byte" {
				t.Errorf("Expected 'character' type to be 'byte', got '%s'", addr.Type)
			}
		}
	})

	t.Run("successful load - jp_1.0.0.4", func(t *testing.T) {
		cfg, err := config.LoadVersionConfigByName("jp_1.0.0.4")
		if err != nil {
			t.Fatalf("LoadVersionConfigByName failed: %v", err)
		}

		// Verify metadata
		if cfg.Metadata.Name != "jp_1.0.0.4" {
			t.Errorf("Expected name 'jp_1.0.0.4', got '%s'", cfg.Metadata.Name)
		}
		if cfg.Metadata.ProcessName != "YSO_WIN.exe" {
			t.Errorf("Expected process_name 'YSO_WIN.exe', got '%s'", cfg.Metadata.ProcessName)
		}
	})

	t.Run("non-existent version", func(t *testing.T) {
		_, err := config.LoadVersionConfigByName("non_existent")
		if err == nil {
			t.Errorf("Expected error for non-existent version")
		}
	})

	t.Run("empty version name", func(t *testing.T) {
		_, err := config.LoadVersionConfigByName("")
		if err == nil {
			t.Errorf("Expected error for empty version name")
		}
	})
}

func TestGetAvailableVersions(t *testing.T) {
	versions := config.GetAvailableVersions()

	if len(versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(versions))
	}

	// Check that expected versions exist
	hasJP103 := false
	hasJP104 := false

	for _, v := range versions {
		if v == "jp_1.1.0.3" {
			hasJP103 = true
		}
		if v == "jp_1.0.0.4" {
			hasJP104 = true
		}
	}

	if !hasJP103 {
		t.Errorf("Expected version 'jp_1.1.0.3' to be available")
	}
	if !hasJP104 {
		t.Errorf("Expected version 'jp_1.0.0.4' to be available")
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "valid version - jp_1.1.0.3",
			version: "jp_1.1.0.3",
			want:    true,
		},
		{
			name:    "valid version - jp_1.0.0.4",
			version: "jp_1.0.0.4",
			want:    true,
		},
		{
			name:    "invalid version",
			version: "invalid_version",
			want:    false,
		},
		{
			name:    "empty version",
			version: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.ValidateVersion(tt.version)
			if got != tt.want {
				t.Errorf("ValidateVersion(%s) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}
