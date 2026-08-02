package collector_test

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/na2na/yso-exporter/internal/collector"
	"github.com/na2na/yso-exporter/internal/config"
	"github.com/na2na/yso-exporter/internal/memory"
	memorymocks "github.com/na2na/yso-exporter/internal/memory/mocks"
	"github.com/na2na/yso-exporter/internal/version"
	versionmocks "github.com/na2na/yso-exporter/internal/version/mocks"
)

// findMetric は、指定した名前のメトリクスをスライスから探します。
func findMetric(metrics []collector.Metric, name string) (collector.Metric, bool) {
	for _, m := range metrics {
		if m.Name == name {
			return m, true
		}
	}
	return collector.Metric{}, false
}

func TestNewYs6Collector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want func(*collector.Ys6Collector) bool
	}{
		{
			name: "正常系: 正しくYs6Collectorが作成される",
			want: func(c *collector.Ys6Collector) bool {
				return c != nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			resolver := versionmocks.NewMockAddressResolver(ctrl)
			reader := memorymocks.NewMockMemoryReader(ctrl)

			got := collector.NewYs6Collector(resolver, reader)

			if !tt.want(got) {
				t.Errorf("NewYs6Collector() validation failed")
			}
		})
	}
}

func TestYs6Collector_Collect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resolver  func(ctrl *gomock.Controller) version.AddressResolver
		reader    func(ctrl *gomock.Controller) memory.MemoryReader
		checkFunc func(t *testing.T, metrics []collector.Metric)
	}{
		{
			name: "正常系: player_baseが非ゼロの場合、ys6_loaded=1でプレイヤー構造体系メトリクスも収集される",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress("room_id").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					BaseAdd: "YS6_WIN.exe+177A48",
					Offsets: []uint32{0x118},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).AnyTimes()
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(777), nil)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil)
				m.EXPECT().ReadInt32(uintptr(0x600118)).Return(int32(3), nil)

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				loaded, ok := findMetric(metrics, "ys6_loaded")
				if !ok || loaded.Value != 1 {
					t.Errorf("ys6_loaded = %v (found=%v), want 1", loaded.Value, ok)
				}

				roomID, ok := findMetric(metrics, "ys6_room_id")
				if !ok || roomID.Value != 3 {
					t.Errorf("ys6_room_id = %v (found=%v), want 3", roomID.Value, ok)
				}
			},
		},
		{
			name: "正常系: player_baseが0の場合、ys6_loaded=0でプレイヤー構造体系メトリクスは出力されない",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil)
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(0), nil)

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				loaded, ok := findMetric(metrics, "ys6_loaded")
				if !ok || loaded.Value != 0 {
					t.Errorf("ys6_loaded = %v (found=%v), want 0", loaded.Value, ok)
				}

				if _, ok := findMetric(metrics, "ys6_room_id"); ok {
					t.Errorf("ys6_room_id should not be emitted when not loaded")
				}
			},
		},
		{
			name: "異常系: player_baseの読み取りに失敗した場合、ys6_loaded=0として扱われる",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil)
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(0), errors.New("read failed"))

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				loaded, ok := findMetric(metrics, "ys6_loaded")
				if !ok || loaded.Value != 0 {
					t.Errorf("ys6_loaded = %v (found=%v), want 0", loaded.Value, ok)
				}
			},
		},
		{
			name: "異常系: boss_hpの読み取りに失敗した場合、0が出力される",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress("boss_hp").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+13A648",
					Offsets: []uint32{},
					Type:    "float32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil)
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(1), nil)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+13A648").Return(uintptr(0x2000), nil)
				m.EXPECT().ReadFloat32(uintptr(0x2000)).Return(float32(0), errors.New("read failed"))

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				bossHP, ok := findMetric(metrics, "ys6_boss_hp")
				if !ok || bossHP.Value != 0 {
					t.Errorf("ys6_boss_hp = %v (found=%v), want 0", bossHP.Value, ok)
				}
			},
		},
		{
			name: "異常系: mp_gaugeの読み取りに失敗した場合、メトリクスが出力されない(欠損)",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress("mp_gauge").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177D10",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil)
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(1), nil)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177D10").Return(uintptr(0x2008), nil)
				m.EXPECT().ReadInt32(uintptr(0x2008)).Return(int32(0), errors.New("read failed"))

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				if _, ok := findMetric(metrics, "ys6_mp_gauge"); ok {
					t.Errorf("ys6_mp_gauge should not be emitted on read failure")
				}
			},
		},
		{
			name: "正常系: ポインタチェーン系メトリクス(hp)が正しく収集される",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress("hp").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+27C2C",
					Offsets: []uint32{0x30},
					Type:    "float32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil)
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(1), nil)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+27C2C").Return(uintptr(0x3000), nil)
				m.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x30}).Return(uintptr(0x300030), nil)
				m.EXPECT().ReadFloat32(uintptr(0x300030)).Return(float32(88.5), nil)

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				hp, ok := findMetric(metrics, "ys6_hp")
				if !ok || hp.Value != float64(float32(88.5)) {
					t.Errorf("ys6_hp = %v (found=%v), want 88.5", hp.Value, ok)
				}
			},
		},
		{
			name: "正常系: position系メトリクス(二重ポインタチェーン)が正しく収集される",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress("position_x").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+14B284",
					Offsets: []uint32{0x0, 0x58},
					Type:    "float32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil)
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(1), nil)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+14B284").Return(uintptr(0x4000), nil)
				m.EXPECT().ResolvePointerChain(uintptr(0x4000), []uintptr{0x0, 0x58}).Return(uintptr(0x400058), nil)
				m.EXPECT().ReadFloat32(uintptr(0x400058)).Return(float32(12.25), nil)

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				x, ok := findMetric(metrics, "ys6_position_x")
				if !ok || x.Value != float64(float32(12.25)) {
					t.Errorf("ys6_position_x = %v (found=%v), want 12.25", x.Value, ok)
				}
			},
		},
		{
			name: "正常系: BaseAddと複数オフセットを併用した場合、Offsets[0]で単純加算しOffsets[1:]をポインタチェインとして解決する",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				// BaseAdd + 複数オフセットの組み合わせは実アドレス表には存在しないが、
				// resolveFinalAddress の Offsets[1:] チェイン経路を検証するための仮想メトリクス定義
				m.EXPECT().ResolveAddress("room_id").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					BaseAdd: "YS6_WIN.exe+177A48",
					Offsets: []uint32{0x118, 0x20},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).AnyTimes()
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(1), nil)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil)
				// Offsets[0]=0x118 は P への単純加算(デリファレンスなし)、
				// Offsets[1:]=[0x20] は 0x600118 を起点とした通常のポインタチェインとして解決される
				m.EXPECT().ResolvePointerChain(uintptr(0x600118), []uintptr{0x20}).Return(uintptr(0x700020), nil)
				m.EXPECT().ReadInt32(uintptr(0x700020)).Return(int32(9), nil)

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				roomID, ok := findMetric(metrics, "ys6_room_id")
				if !ok || roomID.Value != 9 {
					t.Errorf("ys6_room_id = %v (found=%v), want 9", roomID.Value, ok)
				}
			},
		},
		{
			name: "正常系: max_hpが常時収集グループとして収集される",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress("max_hp").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+13A5C8",
					Offsets: []uint32{},
					Type:    "float32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil)
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(0), nil)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+13A5C8").Return(uintptr(0x5000), nil)
				m.EXPECT().ReadFloat32(uintptr(0x5000)).Return(float32(300), nil)

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				maxHP, ok := findMetric(metrics, "ys6_max_hp")
				if !ok || maxHP.Value != 300 {
					t.Errorf("ys6_max_hp = %v (found=%v), want 300", maxHP.Value, ok)
				}
			},
		},
		{
			name: "正常系: on_groundが二重ポインタチェーンで収集される",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress("on_ground").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+14B284",
					Offsets: []uint32{0x0, 0x46C},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil)
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(0), nil)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+14B284").Return(uintptr(0x4000), nil)
				m.EXPECT().ResolvePointerChain(uintptr(0x4000), []uintptr{0x0, 0x46C}).Return(uintptr(0x40046C), nil)
				m.EXPECT().ReadInt32(uintptr(0x40046C)).Return(int32(1), nil)

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				onGround, ok := findMetric(metrics, "ys6_on_ground")
				if !ok || onGround.Value != 1 {
					t.Errorf("ys6_on_ground = %v (found=%v), want 1", onGround.Value, ok)
				}
			},
		},
		{
			name: "正常系: equipped_swordがプレイヤー構造体系として収集される",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress("equipped_sword").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					BaseAdd: "YS6_WIN.exe+177A48",
					Offsets: []uint32{0x430},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).AnyTimes()
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(1), nil)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil)
				m.EXPECT().ReadInt32(uintptr(0x600430)).Return(int32(2), nil)

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				sword, ok := findMetric(metrics, "ys6_equipped_sword")
				if !ok || sword.Value != 2 {
					t.Errorf("ys6_equipped_sword = %v (found=%v), want 2", sword.Value, ok)
				}
			},
		},
		{
			name: "正常系: sword_livart_gaugeがプレイヤー構造体系として収集される",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress("sword_livart_gauge").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					BaseAdd: "YS6_WIN.exe+177A48",
					Offsets: []uint32{0x8B0},
					Type:    "float32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).AnyTimes()
				m.EXPECT().ReadInt32(uintptr(0x1000)).Return(int32(1), nil)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil)
				m.EXPECT().ReadFloat32(uintptr(0x6008B0)).Return(float32(1500), nil)

				return m
			},
			checkFunc: func(t *testing.T, metrics []collector.Metric) {
				gauge, ok := findMetric(metrics, "ys6_sword_livart_gauge")
				if !ok || gauge.Value != 1500 {
					t.Errorf("ys6_sword_livart_gauge = %v (found=%v), want 1500", gauge.Value, ok)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			resolver := tt.resolver(ctrl)
			reader := tt.reader(ctrl)
			c := collector.NewYs6Collector(resolver, reader)

			if err := c.Collect(); err != nil {
				t.Fatalf("Collect() unexpected error: %v", err)
			}

			tt.checkFunc(t, c.GetMetrics())
		})
	}
}

func TestYs6Collector_GetMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFunc func(*collector.Ys6Collector) error
		want      []collector.Metric
	}{
		{
			name:      "正常系: Collect()未実行の場合、空のスライスが返される",
			setupFunc: nil,
			want:      []collector.Metric{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			resolver := versionmocks.NewMockAddressResolver(ctrl)
			reader := memorymocks.NewMockMemoryReader(ctrl)
			c := collector.NewYs6Collector(resolver, reader)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(c); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			got := c.GetMetrics()
			if len(got) != len(tt.want) {
				t.Errorf("GetMetrics() length = %d, want %d", len(got), len(tt.want))
			}
		})
	}
}
