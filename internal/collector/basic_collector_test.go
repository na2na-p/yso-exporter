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

func TestNewBasicCollector(t *testing.T) {
	type args struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name string
		args args
		want func(*collector.BasicCollector) bool // 結果の検証関数
	}{
		{
			name: "正常系: 正しくBasicCollectorが作成される",
			args: args{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					return versionmocks.NewMockAddressResolver(ctrl)
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			want: func(bc *collector.BasicCollector) bool {
				return bc != nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.args.resolver(ctrl)
			reader := tt.args.reader(ctrl)

			got := collector.NewBasicCollector(resolver, reader)

			if !tt.want(got) {
				t.Errorf("NewBasicCollector() validation failed")
			}
		})
	}
}

func TestBasicCollector_Collect(t *testing.T) {
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
			name: "正常系: 全ての基本メトリクスが正常に収集される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)

					// character の解決
					m.EXPECT().ResolveAddress("character").Return(&config.AddressConfig{
						Base:    0x1000,
						Offsets: []uint32{0x10, 0x20},
						Type:    "byte",
					}, nil)

					// level の解決
					m.EXPECT().ResolveAddress("level").Return(&config.AddressConfig{
						Base:    0x2000,
						Offsets: []uint32{0x30},
						Type:    "int32",
					}, nil)

					// experience の解決
					m.EXPECT().ResolveAddress("experience").Return(&config.AddressConfig{
						Base:    0x3000,
						Offsets: []uint32{0x40},
						Type:    "int32",
					}, nil)

					// hp の解決
					m.EXPECT().ResolveAddress("hp").Return(&config.AddressConfig{
						Base:    0x4000,
						Offsets: []uint32{0x50},
						Type:    "int32",
					}, nil)

					// mp の解決
					m.EXPECT().ResolveAddress("mp").Return(&config.AddressConfig{
						Base:    0x5000,
						Offsets: []uint32{0x60},
						Type:    "int32",
					}, nil)

					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)

					// character のメモリ読み取り
					m.EXPECT().ParseAndResolveAddress(0x1000).Return(uintptr(0x1000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x1000), []uintptr{0x10, 0x20}).Return(uintptr(0x100000), nil)
					m.EXPECT().ReadByte(uintptr(0x100000)).Return(byte(0), nil) // 0 = Yunica

					// level のメモリ読み取り
					m.EXPECT().ParseAndResolveAddress(0x2000).Return(uintptr(0x2000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x2000), []uintptr{0x30}).Return(uintptr(0x200000), nil)
					m.EXPECT().ReadInt32(uintptr(0x200000)).Return(int32(15), nil)

					// exp のメモリ読み取り
					m.EXPECT().ParseAndResolveAddress(0x3000).Return(uintptr(0x3000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x40}).Return(uintptr(0x300000), nil)
					m.EXPECT().ReadInt32(uintptr(0x300000)).Return(int32(1234), nil)

					// hp のメモリ読み取り
					m.EXPECT().ParseAndResolveAddress(0x4000).Return(uintptr(0x4000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x4000), []uintptr{0x50}).Return(uintptr(0x400000), nil)
					m.EXPECT().ReadInt32(uintptr(0x400000)).Return(int32(85), nil)

					// mp のメモリ読み取り
					m.EXPECT().ParseAndResolveAddress(0x5000).Return(uintptr(0x5000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x5000), []uintptr{0x60}).Return(uintptr(0x500000), nil)
					m.EXPECT().ReadInt32(uintptr(0x500000)).Return(int32(42), nil)

					return m
				},
			},
			wantErr: nil,
		},
		{
			name: "異常系: アドレス解決に失敗した場合、エラーが返る",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					m.EXPECT().ResolveAddress("character").Return(nil, errors.New("address not found"))
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			wantErr: errors.New("failed to resolve address for character: address not found"),
		},
		{
			name: "異常系: メモリ読み取りに失敗した場合、エラーが返る",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					m.EXPECT().ResolveAddress("character").Return(&config.AddressConfig{
						Base:    0x1000,
						Offsets: []uint32{0x10, 0x20},
						Type:    "byte",
					}, nil)
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					m.EXPECT().ParseAndResolveAddress(0x1000).Return(uintptr(0x1000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x1000), []uintptr{0x10, 0x20}).Return(uintptr(0), errors.New("memory read failed"))
					return m
				},
			},
			wantErr: errors.New("failed to resolve pointer chain for character: memory read failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)
			bc := collector.NewBasicCollector(resolver, reader)

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

func TestBasicCollector_GetMetrics(t *testing.T) {
	type fields struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name      string
		fields    fields
		setupFunc func(*collector.BasicCollector) error // Collect()を事前に実行するかどうか
		want      []collector.Metric
	}{
		{
			name: "正常系: 収集済みのメトリクスが返される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)

					// character の解決
					m.EXPECT().ResolveAddress("character").Return(&config.AddressConfig{
						Base:    0x1000,
						Offsets: []uint32{0x10, 0x20},
						Type:    "byte",
					}, nil)

					// level の解決
					m.EXPECT().ResolveAddress("level").Return(&config.AddressConfig{
						Base:    0x2000,
						Offsets: []uint32{0x30},
						Type:    "int32",
					}, nil)

					// experience の解決
					m.EXPECT().ResolveAddress("experience").Return(&config.AddressConfig{
						Base:    0x3000,
						Offsets: []uint32{0x40},
						Type:    "int32",
					}, nil)

					// hp の解決
					m.EXPECT().ResolveAddress("hp").Return(&config.AddressConfig{
						Base:    0x4000,
						Offsets: []uint32{0x50},
						Type:    "int32",
					}, nil)

					// mp の解決
					m.EXPECT().ResolveAddress("mp").Return(&config.AddressConfig{
						Base:    0x5000,
						Offsets: []uint32{0x60},
						Type:    "int32",
					}, nil)

					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)

					// character のメモリ読み取り
					m.EXPECT().ParseAndResolveAddress(0x1000).Return(uintptr(0x1000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x1000), []uintptr{0x10, 0x20}).Return(uintptr(0x100000), nil)
					m.EXPECT().ReadByte(uintptr(0x100000)).Return(byte(1), nil) // 1 = Hugo

					// level のメモリ読み取り
					m.EXPECT().ParseAndResolveAddress(0x2000).Return(uintptr(0x2000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x2000), []uintptr{0x30}).Return(uintptr(0x200000), nil)
					m.EXPECT().ReadInt32(uintptr(0x200000)).Return(int32(25), nil)

					// exp のメモリ読み取り
					m.EXPECT().ParseAndResolveAddress(0x3000).Return(uintptr(0x3000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x40}).Return(uintptr(0x300000), nil)
					m.EXPECT().ReadInt32(uintptr(0x300000)).Return(int32(5678), nil)

					// hp のメモリ読み取り
					m.EXPECT().ParseAndResolveAddress(0x4000).Return(uintptr(0x4000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x4000), []uintptr{0x50}).Return(uintptr(0x400000), nil)
					m.EXPECT().ReadInt32(uintptr(0x400000)).Return(int32(95), nil)

					// mp のメモリ読み取り
					m.EXPECT().ParseAndResolveAddress(0x5000).Return(uintptr(0x5000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x5000), []uintptr{0x60}).Return(uintptr(0x500000), nil)
					m.EXPECT().ReadInt32(uintptr(0x500000)).Return(int32(68), nil)

					return m
				},
			},
			setupFunc: func(bc *collector.BasicCollector) error {
				return bc.Collect() // メトリクスを事前に収集
			},
			want: []collector.Metric{
				{
					Name:   "character",
					Value:  1, // Hugo
					Type:   collector.Gauge,
					Labels: map[string]string{"character": "hugo"},
				},
				{
					Name:   "level",
					Value:  25,
					Type:   collector.Gauge,
					Labels: map[string]string{},
				},
				{
					Name:   "exp",
					Value:  5678,
					Type:   collector.Gauge,
					Labels: map[string]string{},
				},
				{
					Name:   "hp",
					Value:  95,
					Type:   collector.Gauge,
					Labels: map[string]string{},
				},
				{
					Name:   "mp",
					Value:  68,
					Type:   collector.Gauge,
					Labels: map[string]string{},
				},
			},
		},
		{
			name: "正常系: Collect()未実行の場合、空のスライスが返される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					return versionmocks.NewMockAddressResolver(ctrl)
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			setupFunc: nil, // Collect()を実行しない
			want:      []collector.Metric{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)
			bc := collector.NewBasicCollector(resolver, reader)

			// 必要に応じてCollect()を事前実行
			if tt.setupFunc != nil {
				if err := tt.setupFunc(bc); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			got := bc.GetMetrics()

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetMetrics() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
