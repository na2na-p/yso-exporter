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

func TestNewFlagCollector(t *testing.T) {
	type args struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name string
		args args
		want func(*collector.FlagCollector) bool
	}{
		{
			name: "正常系: 正しくFlagCollectorが作成される",
			args: args{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					return versionmocks.NewMockAddressResolver(ctrl)
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			want: func(fc *collector.FlagCollector) bool {
				return fc != nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.args.resolver(ctrl)
			reader := tt.args.reader(ctrl)

			got := collector.NewFlagCollector(resolver, reader)
			if !tt.want(got) {
				t.Errorf("NewFlagCollector() validation failed")
			}
		})
	}
}

func TestFlagCollector_Collect(t *testing.T) {
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
			name: "正常系: フラグメトリクスが正常に収集される（バイト配列フラグ）",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					
					// story_flags の設定（バイト配列）
					m.EXPECT().ResolveAddress("story_flags").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x30, 0x10},
						Type:    "byte_array",
					}, nil)
					
					// teleport_flags の設定（バイト配列）
					m.EXPECT().ResolveAddress("teleport_flags").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x30, 0x20},
						Type:    "byte_array",
					}, nil)
					
					// event_flags の設定（バイト配列）
					m.EXPECT().ResolveAddress("event_flags").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x30, 0x30},
						Type:    "byte_array",
					}, nil)
					
					// item_flags の設定（バイト配列）
					m.EXPECT().ResolveAddress("item_flags").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x30, 0x40},
						Type:    "byte_array",
					}, nil)
					
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					
					// ParseAndResolveAddress for each base address
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil).Times(4)
					
					// ポインタチェインの解決
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x30, 0x10}).Return(uintptr(0x700000), nil) // story_flags
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x30, 0x20}).Return(uintptr(0x700100), nil) // teleport_flags
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x30, 0x30}).Return(uintptr(0x700200), nil) // event_flags
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x30, 0x40}).Return(uintptr(0x700300), nil) // item_flags
					
					// story_flags のメモリ読み取り (10バイト = 80フラグ)
					// 最初の数バイトのみテスト用データを設定
					m.EXPECT().ReadByte(uintptr(0x700000)).Return(byte(0x05), nil) // バイト0: ビット0,2がセット
					m.EXPECT().ReadByte(uintptr(0x700001)).Return(byte(0x00), nil) // バイト1: 全ビットクリア
					// 残りの8バイトは0として読み取り
					m.EXPECT().ReadByte(gomock.Any()).Return(byte(0x00), nil).Times(8)
					
					// teleport_flags のメモリ読み取り (5バイト = 40フラグ)
					m.EXPECT().ReadByte(gomock.Any()).Return(byte(0x01), nil).Times(5) // 全て0x01
					
					// event_flags のメモリ読み取り (15バイト = 120フラグ)
					m.EXPECT().ReadByte(gomock.Any()).Return(byte(0x00), nil).Times(15)
					
					// item_flags のメモリ読み取り (20バイト = 160フラグ)
					m.EXPECT().ReadByte(gomock.Any()).Return(byte(0x00), nil).Times(20)
					
					return m
				},
			},
			wantErr: nil,
		},
		{
			name: "正常系: 単一フラグメトリクスが正常に収集される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					
					// story_flags の設定（単一フラグ）
					m.EXPECT().ResolveAddress("story_flags").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x30, 0x10},
						Type:    "int32",
					}, nil)
					
					// 他のフラグは設定しない（テストを簡素化）
					m.EXPECT().ResolveAddress("teleport_flags").Return(nil, errors.New("not configured"))
					
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					
					// ParseAndResolveAddress for base address
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil)
					
					// ポインタチェインの解決
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x30, 0x10}).Return(uintptr(0x700000), nil)
					
					// story_flags のメモリ読み取り
					m.EXPECT().ReadInt32(uintptr(0x700000)).Return(int32(12345), nil)
					
					return m
				},
			},
			wantErr: errors.New("failed to resolve address for teleport_flags: not configured"),
		},
		{
			name: "異常系: アドレス解決エラー",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					m.EXPECT().ResolveAddress("story_flags").Return(nil, errors.New("flag address not found"))
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					return memorymocks.NewMockMemoryReader(ctrl)
				},
			},
			wantErr: errors.New("failed to resolve address for story_flags: flag address not found"),
		},
		{
			name: "異常系: ポインタチェイン解決エラー",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					m.EXPECT().ResolveAddress("story_flags").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x30, 0x10},
						Type:    "byte_array",
					}, nil)
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x30, 0x10}).Return(uintptr(0), errors.New("invalid flag pointer"))
					return m
				},
			},
			wantErr: errors.New("failed to resolve pointer chain for story_flags: invalid flag pointer"),
		},
		{
			name: "異常系: フラグメモリ読み取りエラー",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					m.EXPECT().ResolveAddress("story_flags").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x30, 0x10},
						Type:    "byte_array",
					}, nil)
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x30, 0x10}).Return(uintptr(0x700000), nil)
					m.EXPECT().ReadByte(uintptr(0x700000)).Return(byte(0), errors.New("flag read error"))
					return m
				},
			},
			wantErr: errors.New("failed to collect flag metrics for story_flags: failed to read flag byte at offset 0: flag read error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			fc := collector.NewFlagCollector(resolver, reader)
			err := fc.Collect()

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

func TestFlagCollector_GetMetrics(t *testing.T) {
	type fields struct {
		resolver func(ctrl *gomock.Controller) version.AddressResolver
		reader   func(ctrl *gomock.Controller) memory.MemoryReader
	}
	tests := []struct {
		name   string
		fields fields
		want   int // 期待されるメトリクス数
	}{
		{
			name: "正常系: バイト配列フラグからメトリクスが正しく生成される",
			fields: fields{
				resolver: func(ctrl *gomock.Controller) version.AddressResolver {
					m := versionmocks.NewMockAddressResolver(ctrl)
					
					// story_flags のみテスト（簡素化）
					m.EXPECT().ResolveAddress("story_flags").Return(&config.AddressConfig{
						Base:    0x400000,
						Offsets: []uint32{0x30, 0x10},
						Type:    "byte_array",
					}, nil)
					
					// 他のフラグは省略
					m.EXPECT().ResolveAddress("teleport_flags").Return(nil, errors.New("skip"))
					
					return m
				},
				reader: func(ctrl *gomock.Controller) memory.MemoryReader {
					m := memorymocks.NewMockMemoryReader(ctrl)
					
					m.EXPECT().ParseAndResolveAddress(0x400000).Return(uintptr(0x400000), nil)
					m.EXPECT().ResolvePointerChain(uintptr(0x400000), []uintptr{0x30, 0x10}).Return(uintptr(0x700000), nil)
					
					// 10バイトのstory_flags（テスト用に最初の2バイトのみ非ゼロ）
					m.EXPECT().ReadByte(uintptr(0x700000)).Return(byte(0x03), nil) // ビット0,1がセット
					m.EXPECT().ReadByte(uintptr(0x700001)).Return(byte(0x04), nil) // ビット2がセット
					m.EXPECT().ReadByte(gomock.Any()).Return(byte(0x00), nil).Times(8) // 残り8バイト
					
					return m
				},
			},
			want: 80, // story_flags: 10バイト × 8ビット = 80メトリクス
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			resolver := tt.fields.resolver(ctrl)
			reader := tt.fields.reader(ctrl)

			fc := collector.NewFlagCollector(resolver, reader)
			
			// まずCollectを実行（エラーは無視 - 部分的なデータ収集をテスト）
			_ = fc.Collect()

			got := fc.GetMetrics()
			
			// メトリクス数を検証
			if len(got) != tt.want {
				t.Errorf("GetMetrics() returned %d metrics, want %d", len(got), tt.want)
			}
			
			// 最初のいくつかのメトリクスの内容を検証
			if len(got) > 0 {
				// 最初のメトリクス（story_flags_0）
				firstMetric := got[0]
				if firstMetric.Name != "story_flags_0" {
					t.Errorf("First metric name = %s, want story_flags_0", firstMetric.Name)
				}
				if firstMetric.Type != collector.Gauge {
					t.Errorf("First metric type = %v, want Gauge", firstMetric.Type)
				}
				if firstMetric.Labels["flag_type"] != "story_flags" {
					t.Errorf("First metric flag_type label = %s, want story_flags", firstMetric.Labels["flag_type"])
				}
			}
		})
	}
}