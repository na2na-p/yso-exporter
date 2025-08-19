package collector_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.uber.org/mock/gomock"

	"github.com/na2na/yso-exporter/internal/collector"
	"github.com/na2na/yso-exporter/internal/config"
	"github.com/na2na/yso-exporter/internal/memory"
	memorymocks "github.com/na2na/yso-exporter/internal/memory/mocks"
	"github.com/na2na/yso-exporter/internal/version"
	versionmocks "github.com/na2na/yso-exporter/internal/version/mocks"
)

func TestNewPositionCollector(t *testing.T) {
	type args struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name string
		args args
		want func(*collector.PositionCollector) bool // 結果の検証関数
	}{
		{
			name: "正常系: 正しくPositionCollectorが作成される",
			args: args{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					return versionmocks.NewMockAddressResolver(ctrl)
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			want: func(pc *collector.PositionCollector) bool {
				return pc != nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.args.resolver(ctrl)
			reader := tt.args.reader(ctrl)

			got := collector.NewPositionCollector(resolver, reader)
			if !tt.want(got) {
				t.Errorf("NewPositionCollector() validation failed")
			}
		})
	}
}

func TestPositionCollector_Collect(t *testing.T) {
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
			name: "正常系: 位置メトリクスが正常に収集される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					
					// position_x の設定
					m.EXPECT().ResolveAddress("position_x").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x20},
						Type:    "float32",
					}, nil)
					
					// position_y の設定
					m.EXPECT().ResolveAddress("position_y").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x24},
						Type:    "float32",
					}, nil)
					
					// current_map の設定
					m.EXPECT().ResolveAddress("current_map").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x30},
						Type:    "byte",
					}, nil)
					
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					
					// ParseAndResolveAddress for each base address
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil).Times(3)
					
					// ポインタチェインの解決
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x20}).Return(uintptr(0x500000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x24}).Return(uintptr(0x500004), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x30}).Return(uintptr(0x500010), nil)
					
					// メモリ読み取り
					m.EXPECT().ReadFloat32(uintptr(0x500000)).Return(float32(100.5), nil)
					m.EXPECT().ReadFloat32(uintptr(0x500004)).Return(float32(200.3), nil)
					m.EXPECT().ReadByte(uintptr(0x500010)).Return(byte(1), nil)
					
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
					m.EXPECT().ResolveAddress("position_x").Return(nil, errors.New("address not found"))
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			wantErr: errors.New("failed to resolve address for position_x: address not found"),
		},
		{
			name: "異常系: ポインタチェイン解決エラー",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					m.EXPECT().ResolveAddress("position_x").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x20},
						Type:    "float32",
					}, nil)
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x20}).Return(uintptr(0), errors.New("invalid pointer"))
					return m
				},
			},
			wantErr: errors.New("failed to resolve pointer chain for position_x: invalid pointer"),
		},
		{
			name: "異常系: メモリ読み取りエラー",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					m.EXPECT().ResolveAddress("position_x").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x20},
						Type:    "float32",
					}, nil)
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x20}).Return(uintptr(0x500000), nil)
					m.EXPECT().ReadFloat32(uintptr(0x500000)).Return(float32(0), errors.New("read error"))
					return m
				},
			},
			wantErr: errors.New("failed to read metric position_x: failed to read float32: read error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			pc := collector.NewPositionCollector(resolver, reader)
			err := pc.Collect()

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

func TestPositionCollector_GetMetrics(t *testing.T) {
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
			name: "正常系: 収集したメトリクスが返される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					
					// position_x の設定
					m.EXPECT().ResolveAddress("position_x").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x20},
						Type:    "float32",
					}, nil)
					
					// position_y の設定
					m.EXPECT().ResolveAddress("position_y").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x24},
						Type:    "float32",
					}, nil)
					
					// current_map の設定
					m.EXPECT().ResolveAddress("current_map").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x10, 0x30},
						Type:    "byte",
					}, nil)
					
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					
					// ParseAndResolveAddress for each base address
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil).Times(3)
					
					// ポインタチェインの解決
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x20}).Return(uintptr(0x500000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x24}).Return(uintptr(0x500004), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x10, 0x30}).Return(uintptr(0x500010), nil)
					
					// メモリ読み取り
					m.EXPECT().ReadFloat32(uintptr(0x500000)).Return(float32(100.5), nil)
					m.EXPECT().ReadFloat32(uintptr(0x500004)).Return(float32(200.3), nil)
					m.EXPECT().ReadByte(uintptr(0x500010)).Return(byte(2), nil)
					
					return m
				},
			},
			want: []collector.Metric{
				{
					Name:   "position_x",
					Value:  100.5,
					Type:   collector.Gauge,
					Labels: map[string]string{},
				},
				{
					Name:   "position_y",
					Value:  200.3,
					Type:   collector.Gauge,
					Labels: map[string]string{},
				},
				{
					Name:   "current_map",
					Value:  2,
					Type:   collector.Gauge,
					Labels: map[string]string{"map": "darm_tower_3f"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			pc := collector.NewPositionCollector(resolver, reader)
			
			// まずCollectを実行
			err := pc.Collect()
			if err != nil {
				t.Fatalf("Collect() failed: %v", err)
			}

			got := pc.GetMetrics()
			opt := cmpopts.EquateApprox(0.0001, 0) // 浮動小数点の誤差を許容
			if diff := cmp.Diff(tt.want, got, opt); diff != "" {
				t.Errorf("GetMetrics() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}