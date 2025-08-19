package version_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/na2na/yso-exporter/internal/config"
	"github.com/na2na/yso-exporter/internal/version"
	"github.com/google/go-cmp/cmp"
)

func TestNewVersionManager(t *testing.T) {
	type args struct {
		configDir string
	}
	tests := []struct {
		name    string
		args    args
		want    *version.VersionManager
		wantErr error
	}{
		{
			name: "正常系: 有効な設定ディレクトリでVersionManagerが作成される",
			args: args{
				configDir: "../../configs/versions",
			},
			want:    &version.VersionManager{},
			wantErr: nil,
		},
		{
			name: "異常系: 存在しないディレクトリの場合、エラーが返る",
			args: args{
				configDir: "/nonexistent/path",
			},
			want:    nil,
			wantErr: errors.New("config directory not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := version.NewVersionManager(tt.args.configDir)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("NewVersionManager() error = nil, wantErr %v", tt.wantErr)
				}
				// Note: exact error comparison will be implemented with actual errors
				return
			}

			if err != nil {
				t.Fatalf("NewVersionManager() unexpected error: %v", err)
			}

			if got == nil {
				t.Errorf("NewVersionManager() got nil, want non-nil VersionManager")
			}
		})
	}
}

func TestVersionManager_ListAvailableVersions(t *testing.T) {
	type fields struct {
		configDir string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr error
	}{
		{
			name: "正常系: 利用可能なバージョンのリストが返される",
			fields: fields{
				configDir: "../../configs/versions",
			},
			want:    []string{"jp_1.1.0.3"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, err := version.NewVersionManager(tt.fields.configDir)
			if err != nil {
				t.Fatalf("Failed to create VersionManager: %v", err)
			}

			got, err := vm.ListAvailableVersions()

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("ListAvailableVersions() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ListAvailableVersions() unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ListAvailableVersions() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestVersionManager_ValidateVersion(t *testing.T) {
	type fields struct {
		configDir string
	}
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "正常系: 有効なバージョンでエラーが返らない",
			fields: fields{
				configDir: "../../configs/versions",
			},
			args: args{
				version: "jp_1.1.0.3",
			},
			wantErr: nil,
		},
		{
			name: "異常系: 存在しないバージョンでエラーが返る",
			fields: fields{
				configDir: "../../configs/versions",
			},
			args: args{
				version: "nonexistent_version",
			},
			wantErr: errors.New("version not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, err := version.NewVersionManager(tt.fields.configDir)
			if err != nil {
				t.Fatalf("Failed to create VersionManager: %v", err)
			}

			err = vm.ValidateVersion(tt.args.version)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("ValidateVersion() error = nil, wantErr %v", tt.wantErr)
				}
				// Note: exact error comparison will be implemented with actual errors
			} else {
				if err != nil {
					t.Fatalf("ValidateVersion() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestVersionManager_GetVersionPath(t *testing.T) {
	type fields struct {
		configDir string
	}
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr error
	}{
		{
			name: "正常系: 有効なバージョンでパスが返される",
			fields: fields{
				configDir: "../../configs/versions",
			},
			args: args{
				version: "jp_1.1.0.3",
			},
			want:    filepath.Join("../../configs/versions", "jp_1.1.0.3.yaml"),
			wantErr: nil,
		},
		{
			name: "異常系: 存在しないバージョンでエラーが返る",
			fields: fields{
				configDir: "../../configs/versions",
			},
			args: args{
				version: "nonexistent_version",
			},
			want:    "",
			wantErr: errors.New("version not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, err := version.NewVersionManager(tt.fields.configDir)
			if err != nil {
				t.Fatalf("Failed to create VersionManager: %v", err)
			}

			got, err := vm.GetVersionPath(tt.args.version)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("GetVersionPath() error = nil, wantErr %v", tt.wantErr)
				}
				// Note: exact error comparison will be implemented with actual errors
				return
			}

			if err != nil {
				t.Fatalf("GetVersionPath() unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetVersionPath() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestVersionManager_LoadVersionConfig(t *testing.T) {
	type fields struct {
		configDir string
	}
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *config.VersionConfig
		wantErr error
	}{
		{
			name: "正常系: 有効なバージョンで設定が読み込まれる",
			fields: fields{
				configDir: "../../configs/versions",
			},
			args: args{
				version: "jp_1.1.0.3",
			},
			want: &config.VersionConfig{
				// Expected structure will be validated in actual implementation
			},
			wantErr: nil,
		},
		{
			name: "異常系: 存在しないバージョンでエラーが返る",
			fields: fields{
				configDir: "../../configs/versions",
			},
			args: args{
				version: "nonexistent_version",
			},
			want:    nil,
			wantErr: errors.New("version not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, err := version.NewVersionManager(tt.fields.configDir)
			if err != nil {
				t.Fatalf("Failed to create VersionManager: %v", err)
			}

			got, err := vm.LoadVersionConfig(tt.args.version)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("LoadVersionConfig() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("LoadVersionConfig() unexpected error: %v", err)
			}

			if got == nil {
				t.Errorf("LoadVersionConfig() got nil, want non-nil VersionConfig")
			}

			// Validate that the loaded config has expected metadata
			if got.Metadata.Name == "" {
				t.Errorf("LoadVersionConfig() got empty metadata.name")
			}
		})
	}
}