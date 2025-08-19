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

func TestNewCombatCollector(t *testing.T) {
	type args struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name string
		args args
		want func(*collector.CombatCollector) bool
	}{
		{
			name: "正常系: 正しくCombatCollectorが作成される",
			args: args{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					return versionmocks.NewMockAddressResolver(ctrl)
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			want: func(cc *collector.CombatCollector) bool {
				return cc != nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.args.resolver(ctrl)
			reader := tt.args.reader(ctrl)

			got := collector.NewCombatCollector(resolver, reader)
			if !tt.want(got) {
				t.Errorf("NewCombatCollector() validation failed")
			}
		})
	}
}

func TestCombatCollector_Collect(t *testing.T) {
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
			name: "正常系: 戦闘メトリクスが正常に収集される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					
					// current_hp の設定
					m.EXPECT().ResolveAddress("current_hp").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x20},
						Type:    "int32",
					}, nil)
					
					// current_mp の設定
					m.EXPECT().ResolveAddress("current_mp").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x24},
						Type:    "int32",
					}, nil)
					
					// max_hp の設定
					m.EXPECT().ResolveAddress("max_hp").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x28},
						Type:    "int32",
					}, nil)
					
					// max_mp の設定
					m.EXPECT().ResolveAddress("max_mp").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x2C},
						Type:    "int32",
					}, nil)
					
					// attack の設定
					m.EXPECT().ResolveAddress("attack").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x30},
						Type:    "int16",
					}, nil)
					
					// defense の設定
					m.EXPECT().ResolveAddress("defense").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x34},
						Type:    "int16",
					}, nil)
					
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					
					// ParseAndResolveAddress for each base address
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil).Times(6)
					
					// ポインタチェインの解決
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x20}).Return(uintptr(0x500000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x24}).Return(uintptr(0x500004), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x28}).Return(uintptr(0x500008), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x2C}).Return(uintptr(0x50000C), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x30}).Return(uintptr(0x500010), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x34}).Return(uintptr(0x500014), nil)
					
					// メモリ読み取り
					m.EXPECT().ReadInt32(uintptr(0x500000)).Return(int32(250), nil)
					m.EXPECT().ReadInt32(uintptr(0x500004)).Return(int32(80), nil)
					m.EXPECT().ReadInt32(uintptr(0x500008)).Return(int32(300), nil)
					m.EXPECT().ReadInt32(uintptr(0x50000C)).Return(int32(100), nil)
					m.EXPECT().ReadInt32(uintptr(0x500010)).Return(int32(45), nil)  // attack as int32
					m.EXPECT().ReadInt32(uintptr(0x500014)).Return(int32(35), nil)  // defense as int32
					
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
					m.EXPECT().ResolveAddress("current_hp").Return(nil, errors.New("address not found"))
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			wantErr: errors.New("failed to resolve address for current_hp: address not found"),
		},
		{
			name: "異常系: メモリ読み取りエラー",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					m.EXPECT().ResolveAddress("current_hp").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x20},
						Type:    "int32",
					}, nil)
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x20}).Return(uintptr(0x500000), nil)
					m.EXPECT().ReadInt32(uintptr(0x500000)).Return(int32(0), errors.New("read error"))
					return m
				},
			},
			wantErr: errors.New("failed to read metric current_hp: failed to read int32: read error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			cc := collector.NewCombatCollector(resolver, reader)
			err := cc.Collect()

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

func TestCombatCollector_GetMetrics(t *testing.T) {
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
			name: "正常系: 戦闘メトリクスが返される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					
					// 必要なメトリクスの設定
					m.EXPECT().ResolveAddress("current_hp").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x10, 0x20}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("current_mp").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x10, 0x24}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("max_hp").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x10, 0x28}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("max_mp").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x10, 0x2C}, Type: "int32",
					}, nil)
					m.EXPECT().ResolveAddress("attack").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x10, 0x30}, Type: "int16",
					}, nil)
					m.EXPECT().ResolveAddress("defense").Return(&config.AddressConfig{
						Base: 0x400000, Offsets: []uint32{0x10, 0x34}, Type: "int16",
					}, nil)
					
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					
					// ParseAndResolveAddress for each base address
					m.EXPECT().ParseAndResolveAddress(gomock.Any()).Return(uintptr(0x400000), nil).Times(6)
					
					// ポインタチェインの解決とメモリ読み取り
					m.EXPECT().ResolvePointerChain(gomock.Any(), gomock.Any()).Return(uintptr(0x500000), nil).Times(6)
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(250), nil)  // current_hp
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(80), nil)   // current_mp
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(300), nil)  // max_hp
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(100), nil)  // max_mp
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(45), nil)   // attack
					m.EXPECT().ReadInt32(gomock.Any()).Return(int32(35), nil)   // defense
					
					return m
				},
			},
			want: []collector.Metric{
				{
					Name:   "current_hp",
					Value:  250,
					Type:   collector.Gauge,
					Labels: map[string]string{"stat_type": "health"},
				},
				{
					Name:   "current_mp",
					Value:  80,
					Type:   collector.Gauge,
					Labels: map[string]string{"stat_type": "mana"},
				},
				{
					Name:   "max_hp",
					Value:  300,
					Type:   collector.Gauge,
					Labels: map[string]string{"stat_type": "health"},
				},
				{
					Name:   "max_mp",
					Value:  100,
					Type:   collector.Gauge,
					Labels: map[string]string{"stat_type": "mana"},
				},
				{
					Name:   "attack",
					Value:  45,
					Type:   collector.Gauge,
					Labels: map[string]string{"stat_type": "offense"},
				},
				{
					Name:   "defense",
					Value:  35,
					Type:   collector.Gauge,
					Labels: map[string]string{"stat_type": "defense"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			cc := collector.NewCombatCollector(resolver, reader)
			
			// まずCollectを実行
			err := cc.Collect()
			if err != nil {
				t.Fatalf("Collect() failed: %v", err)
			}

			got := cc.GetMetrics()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetMetrics() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}