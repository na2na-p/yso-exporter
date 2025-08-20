package version

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/na2na/yso-exporter/internal/config"
)

func TestNewConfigAddressResolverByVersion(t *testing.T) {
	t.Run("successful creation with valid version", func(t *testing.T) {
		resolver, err := NewConfigAddressResolverByVersion("jp_1.1.0.3")
		if err != nil {
			t.Fatalf("NewConfigAddressResolverByVersion failed: %v", err)
		}

		if resolver == nil {
			t.Fatal("Expected resolver to be non-nil")
		}

		// Verify process name
		if resolver.GetProcessName() != "YSO_WIN.exe" {
			t.Errorf("Expected process name 'YSO_WIN.exe', got '%s'", resolver.GetProcessName())
		}

		// Verify version name
		if resolver.GetVersionName() != "jp_1.1.0.3" {
			t.Errorf("Expected version name 'jp_1.1.0.3', got '%s'", resolver.GetVersionName())
		}
	})

	t.Run("creation fails with invalid version", func(t *testing.T) {
		_, err := NewConfigAddressResolverByVersion("invalid_version")
		if err == nil {
			t.Error("Expected error for invalid version")
		}
	})

	t.Run("creation fails with empty version", func(t *testing.T) {
		_, err := NewConfigAddressResolverByVersion("")
		if err == nil {
			t.Error("Expected error for empty version")
		}
	})
}

func TestConfigAddressResolver_ResolveAddress(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		metricName string
		want       *config.AddressConfig
		wantErr    bool
	}{
		{
			name:       "successful resolution - character",
			version:    "jp_1.1.0.3",
			metricName: "character",
			want: &config.AddressConfig{
				Base:        "YSO_WIN.exe+2F8314",
				Offsets:     []uint32{},
				Type:        "byte",
				Description: "Current character ID",
			},
			wantErr: false,
		},
		{
			name:       "successful resolution - hp",
			version:    "jp_1.1.0.3",
			metricName: "hp",
			want: &config.AddressConfig{
				Base:        "YSO_WIN.exe+2F6ED0",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Current HP",
			},
			wantErr: false,
		},
		{
			name:       "resolution fails for non-existent metric",
			version:    "jp_1.1.0.3",
			metricName: "non_existent",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "resolution fails with empty metric name",
			version:    "jp_1.1.0.3",
			metricName: "",
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := NewConfigAddressResolverByVersion(tt.version)
			if err != nil {
				t.Fatalf("Failed to create resolver: %v", err)
			}

			got, err := resolver.ResolveAddress(tt.metricName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("ResolveAddress() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestConfigAddressResolver_GetProcessName(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "jp_1.1.0.3 process name",
			version: "jp_1.1.0.3",
			want:    "YSO_WIN.exe",
		},
		{
			name:    "jp_1.0.0.4 process name",
			version: "jp_1.0.0.4",
			want:    "YSO_WIN.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := NewConfigAddressResolverByVersion(tt.version)
			if err != nil {
				t.Fatalf("Failed to create resolver: %v", err)
			}

			got := resolver.GetProcessName()
			if got != tt.want {
				t.Errorf("GetProcessName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigAddressResolver_GetVersionName(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "jp_1.1.0.3 version name",
			version: "jp_1.1.0.3",
			want:    "jp_1.1.0.3",
		},
		{
			name:    "jp_1.0.0.4 version name",
			version: "jp_1.0.0.4",
			want:    "jp_1.0.0.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := NewConfigAddressResolverByVersion(tt.version)
			if err != nil {
				t.Fatalf("Failed to create resolver: %v", err)
			}

			got := resolver.GetVersionName()
			if got != tt.want {
				t.Errorf("GetVersionName() = %v, want %v", got, tt.want)
			}
		})
	}
}
