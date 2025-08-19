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

func TestNewBossCollector(t *testing.T) {
	type args struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name string
		args args
		want func(*collector.BossCollector) bool
	}{
		{
			name: "正常系: 正しくBossCollectorが作成される",
			args: args{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					return versionmocks.NewMockAddressResolver(ctrl)
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			want: func(bc *collector.BossCollector) bool {
				return bc != nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.args.resolver(ctrl)
			reader := tt.args.reader(ctrl)

			got := collector.NewBossCollector(resolver, reader)
			if !tt.want(got) {
				t.Errorf("NewBossCollector() validation failed")
			}
		})
	}
}

func TestBossCollector_Collect(t *testing.T) {
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
			name: "正常系: ボス戦メトリクスが正常に収集される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					
					// boss_hp の設定
					m.EXPECT().ResolveAddress("boss_hp").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x20, 0x10},
						Type:    "int32",
					}, nil)
					
					// boss_max_hp の設定
					m.EXPECT().ResolveAddress("boss_max_hp").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x20, 0x14},
						Type:    "int32",
					}, nil)
					
					// boss_id の設定
					m.EXPECT().ResolveAddress("boss_id").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x20, 0x18},
						Type:    "byte",
					}, nil)
					
					// in_boss_fight の設定
					m.EXPECT().ResolveAddress("in_boss_fight").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x20, 0x1C},
						Type:    "byte",
					}, nil)
					
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					
					// ParseAndResolveAddress for each base address
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil).Times(4)
					
					// ポインタチェインの解決
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x20, 0x10}).Return(uintptr(0x600000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x20, 0x14}).Return(uintptr(0x600004), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x20, 0x18}).Return(uintptr(0x600008), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x20, 0x1C}).Return(uintptr(0x60000C), nil)
					
					// メモリ読み取り
					m.EXPECT().ReadInt32(uintptr(0x600000)).Return(int32(5000), nil)  // boss_hp
					m.EXPECT().ReadInt32(uintptr(0x600004)).Return(int32(8000), nil)  // boss_max_hp
					m.EXPECT().ReadByte(uintptr(0x600008)).Return(byte(1), nil)       // boss_id (gyalva)
					m.EXPECT().ReadByte(uintptr(0x60000C)).Return(byte(1), nil)       // in_boss_fight
					
					return m
				},
			},
			wantErr: nil,
		},
		{
			name: "異常系: アドレス解決エラー",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					m.EXPECT().ResolveAddress("boss_hp").Return(nil, errors.New("boss address not found"))
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			wantErr: errors.New("failed to resolve address for boss_hp: boss address not found"),
		},
		{
			name: "異常系: ポインタチェイン解決エラー",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					m.EXPECT().ResolveAddress("boss_hp").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x20, 0x10},
						Type:    "int32",
					}, nil)
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x20, 0x10}).Return(uintptr(0), errors.New("invalid boss pointer"))
					return m
				},
			},
			wantErr: errors.New("failed to resolve pointer chain for boss_hp: invalid boss pointer"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			bc := collector.NewBossCollector(resolver, reader)
			err := bc.Collect()

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

func TestBossCollector_GetMetrics(t *testing.T) {
	type fields struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name   string
		fields fields
		want   []collector.Metric
	}{
		{
			name: "正常系: ボス戦メトリクスが返される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					
					// 必要なメトリクスの設定
					m.EXPECT().ResolveAddress("boss_hp").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x20, 0x10}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("boss_max_hp").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x20, 0x14}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("boss_id").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x20, 0x18}, Type: "byte",
					}, nil)
					m.EXPECT().ResolveAddress("in_boss_fight").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x20, 0x1C}, Type: "byte",
					}, nil)
					
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					
					// ParseAndResolveAddress for each base address
					m.EXPECT().ParseAndResolveAddress(gomock.Any()).Return(uintptr(0x400000), nil).Times(4)
					
					// ポインタチェインの解決
					m.EXPECT().ResolvePointerChain(gomock.Any(), gomock.Any()).Return(uintptr(0x600000), nil).Times(4)
					
					// メモリ読み取り
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(3000), nil)  // boss_hp
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(8000), nil)  // boss_max_hp
					m.EXPECT().ReadByte(gomock.Any()).Return(byte(2), nil)       // boss_id (epona)
					m.EXPECT().ReadByte(gomock.Any()).Return(byte(1), nil)       // in_boss_fight
					
					return m
				},
			},
			want: []collector.Metric{
				{
					Name:  "boss_hp",
					Value: 3000,
					Type:  collector.Gauge,
					Labels: map[string]string{
						"metric_type": "health",
					},
				},
				{
					Name:  "boss_max_hp",
					Value: 8000,
					Type:  collector.Gauge,
					Labels: map[string]string{
						"metric_type": "health",
					},
				},
				{
					Name:  "boss_id",
					Value: 2,
					Type:  collector.Gauge,
					Labels: map[string]string{
						"boss":        "epona",
						"metric_type": "identifier",
					},
				},
				{
					Name:  "in_boss_fight",
					Value: 1,
					Type:  collector.Gauge,
					Labels: map[string]string{
						"status":      "boss_fight",
						"metric_type": "state",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			bc := collector.NewBossCollector(resolver, reader)
			
			// まずCollectを実行
			err := bc.Collect()
			if err != nil {
				t.Fatalf("Collect() failed: %v", err)
			}

			got := bc.GetMetrics()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetMetrics() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}