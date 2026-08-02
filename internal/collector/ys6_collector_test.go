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
			name: "正常系: player_baseの合算値が非ゼロの場合、ys6_loaded=1でプレイヤー構造体系メトリクスも収集される",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					BaseAdd: "YS6_WIN.exe+177A48",
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

				// ゲーティング判定とroom_idのアドレス解決の両方で同じBase/BaseAddが解決されるため2回呼ばれる
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).Times(2)
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil).Times(2)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil).Times(2)
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
			name: "正常系: player_baseの合算値が0の場合、ys6_loaded=0でプレイヤー構造体系メトリクスは出力されない",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					BaseAdd: "YS6_WIN.exe+177A48",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil)
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), nil)

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
			name: "異常系: player_baseの合算解決に失敗した場合、ys6_loaded=0として扱われる",
			resolver: func(ctrl *gomock.Controller) version.AddressResolver {
				m := versionmocks.NewMockAddressResolver(ctrl)

				m.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
					Base:    "YS6_WIN.exe+177A44",
					BaseAdd: "YS6_WIN.exe+177A48",
					Offsets: []uint32{},
					Type:    "int32",
				}, nil)

				m.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

				return m
			},
			reader: func(ctrl *gomock.Controller) memory.MemoryReader {
				m := memorymocks.NewMockMemoryReader(ctrl)

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil)
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), errors.New("read failed"))

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
					BaseAdd: "YS6_WIN.exe+177A48",
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
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil)

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
					BaseAdd: "YS6_WIN.exe+177A48",
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
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil)

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
					BaseAdd: "YS6_WIN.exe+177A48",
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
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil)

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
					BaseAdd: "YS6_WIN.exe+177A48",
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
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil)

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
					BaseAdd: "YS6_WIN.exe+177A48",
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

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).Times(2)
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil).Times(2)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil).Times(2)
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
					BaseAdd: "YS6_WIN.exe+177A48",
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
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), nil)

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
					BaseAdd: "YS6_WIN.exe+177A48",
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
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), nil)

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
					BaseAdd: "YS6_WIN.exe+177A48",
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

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).Times(2)
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil).Times(2)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil).Times(2)
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
					BaseAdd: "YS6_WIN.exe+177A48",
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

				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).Times(2)
				m.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil).Times(2)
				m.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil).Times(2)
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

// TestYs6Collector_Collect_MetricRetentionToleratesTwoFailures は、常時収集メトリクス(zeroOnFailure以外)が
// 直前の成功値を最大2連続失敗まで保持し、3連続失敗で欠損に切り替わることを検証します。
func TestYs6Collector_Collect_MetricRetentionToleratesTwoFailures(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	resolver := versionmocks.NewMockAddressResolver(ctrl)
	reader := memorymocks.NewMockMemoryReader(ctrl)

	resolver.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+177A44",
		BaseAdd: "YS6_WIN.exe+177A48",
		Offsets: []uint32{},
		Type:    "int32",
	}, nil).AnyTimes()
	resolver.EXPECT().ResolveAddress("hp").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+27C2C",
		Offsets: []uint32{0x30},
		Type:    "float32",
	}, nil).AnyTimes()
	resolver.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).AnyTimes()
	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil).AnyTimes()
	// player_baseは常に合算非ゼロ(ロード中)とし、本物のアンロードによる保持値クリアがhpの検証に混入しないようにする
	reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil).AnyTimes()
	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+27C2C").Return(uintptr(0x3000), nil).AnyTimes()

	reader.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x30}).Return(uintptr(0x300030), nil)
	reader.EXPECT().ReadFloat32(uintptr(0x300030)).Return(float32(88.5), nil)
	reader.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x30}).Return(uintptr(0), errors.New("read failed"))
	reader.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x30}).Return(uintptr(0), errors.New("read failed"))
	reader.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x30}).Return(uintptr(0), errors.New("read failed"))
	reader.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x30}).Return(uintptr(0x300030), nil) // cycle5: 成功で連続失敗数がリセットされる
	reader.EXPECT().ReadFloat32(uintptr(0x300030)).Return(float32(92.0), nil)
	reader.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x30}).Return(uintptr(0), errors.New("read failed")) // cycle6: 単発失敗

	c := collector.NewYs6Collector(resolver, reader)

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle1 Collect() unexpected error: %v", err)
	}
	if hp, ok := findMetric(c.GetMetrics(), "ys6_hp"); !ok || hp.Value != float64(float32(88.5)) {
		t.Fatalf("cycle1 ys6_hp = %v (found=%v), want 88.5", hp.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle2 Collect() unexpected error: %v", err)
	}
	if hp, ok := findMetric(c.GetMetrics(), "ys6_hp"); !ok || hp.Value != float64(float32(88.5)) {
		t.Errorf("cycle2 (1回目失敗) ys6_hp = %v (found=%v), want 88.5 (直前値保持)", hp.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle3 Collect() unexpected error: %v", err)
	}
	if hp, ok := findMetric(c.GetMetrics(), "ys6_hp"); !ok || hp.Value != float64(float32(88.5)) {
		t.Errorf("cycle3 (2回目失敗) ys6_hp = %v (found=%v), want 88.5 (直前値保持)", hp.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle4 Collect() unexpected error: %v", err)
	}
	if _, ok := findMetric(c.GetMetrics(), "ys6_hp"); ok {
		t.Errorf("cycle4 (3回目失敗) ys6_hp should not be emitted (本物の欠測)")
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle5 Collect() unexpected error: %v", err)
	}
	if hp, ok := findMetric(c.GetMetrics(), "ys6_hp"); !ok || hp.Value != float64(float32(92.0)) {
		t.Fatalf("cycle5 (成功) ys6_hp = %v (found=%v), want 92.0", hp.Value, ok)
	}

	// cycle6は失敗1回のみ(連続2回ではない)。cycle5の成功で連続失敗数がリセットされていなければ
	// cycle2〜4で積み上がった失敗数がそのまま持ち越され、許容範囲を超えて欠損してしまう。
	if err := c.Collect(); err != nil {
		t.Fatalf("cycle6 Collect() unexpected error: %v", err)
	}
	if hp, ok := findMetric(c.GetMetrics(), "ys6_hp"); !ok || hp.Value != float64(float32(92.0)) {
		t.Errorf("cycle6 (直後の単発失敗) ys6_hp = %v (found=%v), want 92.0 (連続失敗数はcycle5の成功でリセットされ、保持値が出力されるべき)", hp.Value, ok)
	}
}

// TestYs6Collector_Collect_PlayerBaseUnloadClearsRetention は、player_baseの合算値が実際に0を返した際に
// プレイヤー系メトリクスの保持値がクリアされ、再ロード後の一時的失敗で古い値が漏れ出さないことを検証します。
func TestYs6Collector_Collect_PlayerBaseUnloadClearsRetention(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	resolver := versionmocks.NewMockAddressResolver(ctrl)
	reader := memorymocks.NewMockMemoryReader(ctrl)

	resolver.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+177A44",
		BaseAdd: "YS6_WIN.exe+177A48",
		Offsets: []uint32{},
		Type:    "int32",
	}, nil).AnyTimes()
	resolver.EXPECT().ResolveAddress("room_id").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+177A44",
		BaseAdd: "YS6_WIN.exe+177A48",
		Offsets: []uint32{0x118},
		Type:    "int32",
	}, nil).AnyTimes()
	resolver.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).AnyTimes()
	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil).AnyTimes()

	gomock.InOrder(
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil), // cycle1: ゲーティング判定(ロード中)
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil), // cycle1: room_idのアドレス解決
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), nil),        // cycle2: 本物のアンロード
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil), // cycle3: ゲーティング判定(再ロード)
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil), // cycle3: room_idのアドレス解決
	)

	reader.EXPECT().ReadInt32(uintptr(0x600118)).Return(int32(3), nil)                       // cycle1: room_id成功
	reader.EXPECT().ReadInt32(uintptr(0x600118)).Return(int32(0), errors.New("read failed")) // cycle3: room_id失敗

	c := collector.NewYs6Collector(resolver, reader)

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle1 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 1 {
		t.Fatalf("cycle1 ys6_loaded = %v (found=%v), want 1", loaded.Value, ok)
	}
	if roomID, ok := findMetric(c.GetMetrics(), "ys6_room_id"); !ok || roomID.Value != 3 {
		t.Fatalf("cycle1 ys6_room_id = %v (found=%v), want 3", roomID.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle2 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 0 {
		t.Fatalf("cycle2 ys6_loaded = %v (found=%v), want 0", loaded.Value, ok)
	}
	if _, ok := findMetric(c.GetMetrics(), "ys6_room_id"); ok {
		t.Fatalf("cycle2 (アンロード中) ys6_room_id should not be emitted")
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle3 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 1 {
		t.Fatalf("cycle3 ys6_loaded = %v (found=%v), want 1", loaded.Value, ok)
	}
	if _, ok := findMetric(c.GetMetrics(), "ys6_room_id"); ok {
		t.Errorf("cycle3: アンロード時にクリアされた保持値が再ロード後の失敗時に漏れ出してはならない")
	}
}

// TestYs6Collector_Collect_PlayerBaseTransientErrorsTolerateBeforeGivingUp は、player_baseの合算解決自体が
// 連続で失敗した場合に、2サイクルまでは直前のloaded状態(とプレイヤー系メトリクス収集)を維持し、
// 3サイクル目でloaded=0へ切り替えて保持値もクリアすることを検証します。
func TestYs6Collector_Collect_PlayerBaseTransientErrorsTolerateBeforeGivingUp(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	resolver := versionmocks.NewMockAddressResolver(ctrl)
	reader := memorymocks.NewMockMemoryReader(ctrl)

	resolver.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+177A44",
		BaseAdd: "YS6_WIN.exe+177A48",
		Offsets: []uint32{},
		Type:    "int32",
	}, nil).AnyTimes()
	resolver.EXPECT().ResolveAddress("room_id").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+177A44",
		BaseAdd: "YS6_WIN.exe+177A48",
		Offsets: []uint32{0x118},
		Type:    "int32",
	}, nil).AnyTimes()
	resolver.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).AnyTimes()
	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil).AnyTimes()

	gomock.InOrder(
		// cycle1: ゲーティング成功→room_id解決成功
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil),
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil),
		// cycle2: ゲーティング失敗(1回目)。直前のロード状態を維持しroom_idは通常どおり試行される
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), errors.New("read failed")),
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil),
		// cycle3: ゲーティング失敗(2回目)。直前のロード状態を維持しroom_idは通常どおり試行される
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), errors.New("read failed")),
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil),
		// cycle4: ゲーティング失敗(3回目・諦め)。未ロード扱いとなりroom_idは試行されない
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), errors.New("read failed")),
		// cycle5: ゲーティング復帰→room_idも通常どおり試行(値自体はReadInt32で失敗させ保持値クリア後の欠測を検証)
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil),
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil),
	)

	gomock.InOrder(
		reader.EXPECT().ReadInt32(uintptr(0x600118)).Return(int32(5), nil),                       // cycle1
		reader.EXPECT().ReadInt32(uintptr(0x600118)).Return(int32(6), nil),                       // cycle2(loaded維持で通常どおり試行)
		reader.EXPECT().ReadInt32(uintptr(0x600118)).Return(int32(7), nil),                       // cycle3(loaded維持で通常どおり試行)
		reader.EXPECT().ReadInt32(uintptr(0x600118)).Return(int32(0), errors.New("read failed")), // cycle5(保持値クリア後の失敗)
	)

	c := collector.NewYs6Collector(resolver, reader)

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle1 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 1 {
		t.Fatalf("cycle1 ys6_loaded = %v (found=%v), want 1", loaded.Value, ok)
	}
	if roomID, ok := findMetric(c.GetMetrics(), "ys6_room_id"); !ok || roomID.Value != 5 {
		t.Fatalf("cycle1 ys6_room_id = %v (found=%v), want 5", roomID.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle2 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 1 {
		t.Fatalf("cycle2 (player_base 1回目失敗) ys6_loaded = %v (found=%v), want 1 (維持)", loaded.Value, ok)
	}
	if roomID, ok := findMetric(c.GetMetrics(), "ys6_room_id"); !ok || roomID.Value != 6 {
		t.Fatalf("cycle2 ys6_room_id = %v (found=%v), want 6", roomID.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle3 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 1 {
		t.Fatalf("cycle3 (player_base 2回目失敗) ys6_loaded = %v (found=%v), want 1 (維持)", loaded.Value, ok)
	}
	if roomID, ok := findMetric(c.GetMetrics(), "ys6_room_id"); !ok || roomID.Value != 7 {
		t.Fatalf("cycle3 ys6_room_id = %v (found=%v), want 7", roomID.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle4 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 0 {
		t.Fatalf("cycle4 (player_base 3回目失敗) ys6_loaded = %v (found=%v), want 0 (諦め)", loaded.Value, ok)
	}
	if _, ok := findMetric(c.GetMetrics(), "ys6_room_id"); ok {
		t.Fatalf("cycle4: unloaded扱いのためys6_room_idは収集されない")
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle5 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 1 {
		t.Fatalf("cycle5 (player_base復帰) ys6_loaded = %v (found=%v), want 1", loaded.Value, ok)
	}
	if _, ok := findMetric(c.GetMetrics(), "ys6_room_id"); ok {
		t.Errorf("cycle5: 3回失敗時にクリアされた保持値が復帰後の失敗時に漏れ出してはならない")
	}
}

// TestYs6Collector_Collect_PlayerBaseFailureCounterResetsAfterSuccess は、player_baseの合算解決の
// 連続失敗が諦め閾値に達した後でも、1回成功すれば連続失敗数がリセットされ、
// その直後の単発失敗ではロード状態を維持することを検証します。
// リセットが無いと諦め閾値に達した失敗数がそのまま持ち越され、直後の単発失敗だけで
// 誤ってロード状態が失われてしまいます。
func TestYs6Collector_Collect_PlayerBaseFailureCounterResetsAfterSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	resolver := versionmocks.NewMockAddressResolver(ctrl)
	reader := memorymocks.NewMockMemoryReader(ctrl)

	resolver.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+177A44",
		BaseAdd: "YS6_WIN.exe+177A48",
		Offsets: []uint32{},
		Type:    "int32",
	}, nil).AnyTimes()
	resolver.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).AnyTimes()
	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil).AnyTimes()

	gomock.InOrder(
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), errors.New("read failed")), // cycle1: 1回目失敗
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), errors.New("read failed")), // cycle2: 2回目失敗
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), errors.New("read failed")), // cycle3: 3回目失敗(諦め)
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil),                // cycle4: 成功(失敗数リセット)
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), errors.New("read failed")), // cycle5: 直後の単発失敗
	)

	c := collector.NewYs6Collector(resolver, reader)

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle1 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 0 {
		t.Fatalf("cycle1 ys6_loaded = %v (found=%v), want 0", loaded.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle2 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 0 {
		t.Fatalf("cycle2 ys6_loaded = %v (found=%v), want 0", loaded.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle3 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 0 {
		t.Fatalf("cycle3 (3回目失敗・諦め) ys6_loaded = %v (found=%v), want 0", loaded.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle4 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 1 {
		t.Fatalf("cycle4 (成功) ys6_loaded = %v (found=%v), want 1", loaded.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle5 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 1 {
		t.Errorf("cycle5 (直後の単発失敗) ys6_loaded = %v (found=%v), want 1 (cycle4の成功で失敗数がリセットされ、ロード状態が維持されるべき)", loaded.Value, ok)
	}
}

// TestYs6Collector_Collect_PlayerBaseSumNonZeroLoadsEvenIfWordAIsZero は、player_baseのロード判定が
// Base単体の値ではなく、Base+BaseAddの合算値(ResolveSumBaseの戻り値)で行われることを固定します。
// Base側の値がプレイ中に一時的に0へ変動しても、合算値さえ非ゼロであればロード中と判定され、
// プレイヤー系メトリクスが収集されなければなりません。
// また、合算値自体をアドレスとしてデリファレンスしてはいけないため、ゲーティング用に
// ReadInt32(合算値)を期待するmockは設定していません(誤って呼ばれれば想定外呼び出しとして失敗します)。
func TestYs6Collector_Collect_PlayerBaseSumNonZeroLoadsEvenIfWordAIsZero(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	resolver := versionmocks.NewMockAddressResolver(ctrl)
	reader := memorymocks.NewMockMemoryReader(ctrl)

	resolver.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+177A44",
		BaseAdd: "YS6_WIN.exe+177A48",
		Offsets: []uint32{},
		Type:    "int32",
	}, nil)
	resolver.EXPECT().ResolveAddress("room_id").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+177A44",
		BaseAdd: "YS6_WIN.exe+177A48",
		Offsets: []uint32{0x118},
		Type:    "int32",
	}, nil)
	resolver.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).Times(2)
	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil).Times(2)
	// Base側の値が実質0であっても、BaseAdd側の実体ポインタの加算により合算は非ゼロになりうる
	reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x4E90000), nil).Times(2)
	reader.EXPECT().ReadInt32(uintptr(0x4E90118)).Return(int32(3), nil)

	c := collector.NewYs6Collector(resolver, reader)

	if err := c.Collect(); err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded")
	if !ok || loaded.Value != 1 {
		t.Fatalf("ys6_loaded = %v (found=%v), want 1", loaded.Value, ok)
	}
	roomID, ok := findMetric(c.GetMetrics(), "ys6_room_id")
	if !ok || roomID.Value != 3 {
		t.Errorf("ys6_room_id = %v (found=%v), want 3", roomID.Value, ok)
	}
}

// TestYs6Collector_Collect_PlayerBaseSumZeroClearsRetention は、player_baseの合算値
// (ResolveSumBaseの戻り値)が0に解決された場合にのみ本物のアンロードとして扱われ、
// 常時収集メトリクスの保持値もクリアされることを検証します。
func TestYs6Collector_Collect_PlayerBaseSumZeroClearsRetention(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	resolver := versionmocks.NewMockAddressResolver(ctrl)
	reader := memorymocks.NewMockMemoryReader(ctrl)

	resolver.EXPECT().ResolveAddress("player_base").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+177A44",
		BaseAdd: "YS6_WIN.exe+177A48",
		Offsets: []uint32{},
		Type:    "int32",
	}, nil).AnyTimes()
	resolver.EXPECT().ResolveAddress("hp").Return(&config.AddressConfig{
		Base:    "YS6_WIN.exe+27C2C",
		Offsets: []uint32{0x30},
		Type:    "float32",
	}, nil).AnyTimes()
	resolver.EXPECT().ResolveAddress(gomock.Any()).Return(nil, errors.New("not found")).AnyTimes()

	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A44").Return(uintptr(0x1000), nil).AnyTimes()
	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+177A48").Return(uintptr(0x1004), nil).AnyTimes()
	reader.EXPECT().ParseAndResolveAddress("YS6_WIN.exe+27C2C").Return(uintptr(0x3000), nil).AnyTimes()

	gomock.InOrder(
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil), // cycle1: ロード中
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0), nil),        // cycle2: 合算0=本物のアンロード
		reader.EXPECT().ResolveSumBase(uintptr(0x1000), uintptr(0x1004)).Return(uintptr(0x600000), nil), // cycle3: 再ロード
	)

	gomock.InOrder(
		reader.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x30}).Return(uintptr(0x300030), nil), // cycle1: hp成功
		reader.EXPECT().ReadFloat32(uintptr(0x300030)).Return(float32(88.5), nil),
		reader.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x30}).Return(uintptr(0), errors.New("read failed")), // cycle2: アンロード直後の読み取り失敗
		reader.EXPECT().ResolvePointerChain(uintptr(0x3000), []uintptr{0x30}).Return(uintptr(0), errors.New("read failed")), // cycle3: 再ロード直後の読み取り失敗
	)

	c := collector.NewYs6Collector(resolver, reader)

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle1 Collect() unexpected error: %v", err)
	}
	if hp, ok := findMetric(c.GetMetrics(), "ys6_hp"); !ok || hp.Value != float64(float32(88.5)) {
		t.Fatalf("cycle1 ys6_hp = %v (found=%v), want 88.5", hp.Value, ok)
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle2 Collect() unexpected error: %v", err)
	}
	if loaded, ok := findMetric(c.GetMetrics(), "ys6_loaded"); !ok || loaded.Value != 0 {
		t.Fatalf("cycle2 ys6_loaded = %v (found=%v), want 0", loaded.Value, ok)
	}
	if _, ok := findMetric(c.GetMetrics(), "ys6_hp"); ok {
		t.Fatalf("cycle2 (アンロード直後の読み取り失敗) ys6_hp should not be emitted")
	}

	if err := c.Collect(); err != nil {
		t.Fatalf("cycle3 Collect() unexpected error: %v", err)
	}
	if _, ok := findMetric(c.GetMetrics(), "ys6_hp"); ok {
		t.Errorf("cycle3: 合算0でクリアされた保持値がcycle1の88.5のまま漏れ出してはならない(クリアされていれば保持値なしのため欠測になる)")
	}
}
