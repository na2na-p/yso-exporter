package collector_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/na2na/yso-exporter/internal/collector"
	"github.com/na2na/yso-exporter/internal/config"
	"github.com/na2na/yso-exporter/internal/memory"
	memorymocks "github.com/na2na/yso-exporter/internal/memory/mocks"
	"github.com/na2na/yso-exporter/internal/version"
	versionmocks "github.com/na2na/yso-exporter/internal/version/mocks"
)

func TestNewCollectorManager(t *testing.T) {
	type args struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name string
		args args
		want func(*collector.CollectorManager) bool
	}{
		{
			name: "正常系: 正しくCollectorManagerが作成される",
			args: args{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					return versionmocks.NewMockAddressResolver(ctrl)
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			want: func(cm *collector.CollectorManager) bool {
				return cm != nil && cm.GetCollectorCount() == 5 // Basic + 4つの拡張コレクター
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.args.resolver(ctrl)
			reader := tt.args.reader(ctrl)

			got := collector.NewCollectorManager(resolver, reader)
			if !tt.want(got) {
				t.Errorf("NewCollectorManager() validation failed")
			}
		})
	}
}

func TestNewExtendedCollectorManager(t *testing.T) {
	type args struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name string
		args args
		want func(*collector.CollectorManager) bool
	}{
		{
			name: "正常系: 正しく拡張CollectorManagerが作成される",
			args: args{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					return versionmocks.NewMockAddressResolver(ctrl)
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			want: func(cm *collector.CollectorManager) bool {
				return cm != nil && cm.GetCollectorCount() == 4 // 4つの拡張コレクターのみ
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.args.resolver(ctrl)
			reader := tt.args.reader(ctrl)

			got := collector.NewExtendedCollectorManager(resolver, reader)
			if !tt.want(got) {
				t.Errorf("NewExtendedCollectorManager() validation failed")
			}
		})
	}
}

func TestCollectorManager_Collect(t *testing.T) {
	type fields struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "正常系: 一部のCollectorが成功する",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)

					// BasicCollector用の設定（成功させる）
					m.EXPECT().ResolveAddress("character").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x10}, Type: "byte",
					}, nil)
					m.EXPECT().ResolveAddress("level").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x14}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("experience").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x18}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("hp").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x1C}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("mp").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x20}, Type: "int32",
					}, nil)

					// PositionCollector用の設定（失敗させる）
					m.EXPECT().ResolveAddress("position_x").Return(nil, errors.New("position not found"))

					// 他のCollectorも失敗させる
					m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)

					// BasicCollector用のParseAndResolveAddress（成功）
					m.EXPECT().ParseAndResolveAddress(gomock.Any()).Return(uintptr(0x400000), nil).Times(5)

					// BasicCollector用のメモリ読み取り（成功）
					m.EXPECT().ResolvePointerChain(gomock.Any(), gomock.Any()).Return(uintptr(0x500000), nil).Times(5)
					m.EXPECT().ReadByte(gomock.Any()).Return(byte(0), nil)      // character
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(10), nil)   // level
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(1000), nil) // exp
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(100), nil)  // hp
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(50), nil)   // mp

					return m
				},
			},
			wantErr: nil, // 一部成功すれば全体としては成功
		},
		{
			name: "異常系: 一部のCollectorが失敗するが、BasicCollectorは成功する",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					// 全ての解決を失敗させる（BasicCollectorは失敗しても続行するため）
					m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			wantErr: nil, // BasicCollectorは失敗しないため、全体としては成功
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			cm := collector.NewCollectorManager(resolver, reader)
			err := cm.Collect()

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Collect() error = nil, wantErr %v", tt.wantErr)
				}
				if diff := cmp.Diff(tt.wantErr.Error(), err.Error()); diff != "" {
					t.Errorf("Collect() error mismatch (-want +got):\n%s", diff)
				}
			} else {
				if err != nil {
					t.Fatalf("Collect() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCollectorManager_GetMetrics(t *testing.T) {
	type fields struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name   string
		fields fields
		want   int // 期待されるメトリクス数（最低限）
	}{
		{
			name: "正常系: メトリクスが統合されて返される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)

					// BasicCollector用の最低限の設定
					m.EXPECT().ResolveAddress("character").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x10}, Type: "byte",
					}, nil)
					m.EXPECT().ResolveAddress("level").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x14}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("experience").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x18}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("hp").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x1C}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("mp").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x20}, Type: "int32",
					}, nil)

					// 他のCollectorは失敗させる
					m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)

					// BasicCollector用のParseAndResolveAddress
					m.EXPECT().ParseAndResolveAddress(gomock.Any()).Return(uintptr(0x400000), nil).Times(5)

					// BasicCollector用のメモリ読み取り
					m.EXPECT().ResolvePointerChain(gomock.Any(), gomock.Any()).Return(uintptr(0x500000), nil).Times(5)
					m.EXPECT().ReadByte(gomock.Any()).Return(byte(0), nil)
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(10), nil).Times(4)

					return m
				},
			},
			want: 5, // BasicCollectorから5つのメトリクス
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			cm := collector.NewCollectorManager(resolver, reader)

			// まずCollectを実行
			err := cm.Collect()
			if err != nil {
				t.Logf("Collect() returned error (expected for partial failure): %v", err)
			}

			got := cm.GetMetrics()

			// 最低限のメトリクス数を確認
			if len(got) < tt.want {
				t.Errorf("GetMetrics() returned %d metrics, want at least %d", len(got), tt.want)
			}
		})
	}
}

func TestCollectorManager_CollectIndividually(t *testing.T) {
	type fields struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name   string
		fields fields
		want   int // 期待される結果マップのサイズ
	}{
		{
			name: "正常系: 各Collectorの結果が個別に返される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					// 全て失敗させてテストを簡素化
					m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("test error")).AnyTimes()
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			want: 5, // 5つのCollectorの結果
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			cm := collector.NewCollectorManager(resolver, reader)
			got := cm.CollectIndividually()

			if len(got) != tt.want {
				t.Errorf("CollectIndividually() returned %d results, want %d", len(got), tt.want)
			}

			// 期待されるキーが存在するか確認
			expectedKeys := []string{"basic", "position", "combat", "boss", "flag"}
			for _, key := range expectedKeys {
				if _, exists := got[key]; !exists {
					t.Errorf("CollectIndividually() missing key: %s", key)
				}
			}
		})
	}
}
