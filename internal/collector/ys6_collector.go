package collector

import (
	"fmt"

	"github.com/na2na/yso-exporter/internal/config"
	"github.com/na2na/yso-exporter/internal/memory"
	"github.com/na2na/yso-exporter/internal/version"
)

// コンパイル時にYs6CollectorがCollectorインターフェースを実装していることを確認
var _ Collector = (*Ys6Collector)(nil)

// ys6ZeroOnFailureMetrics は、読み取り失敗時に欠損ではなく0を出力するメトリクスのキー集合です。
// ボスが存在しない場面では読み取れないことがあるため、0出力とします。
var ys6ZeroOnFailureMetrics = map[string]bool{
	"boss_hp":     true,
	"boss_max_hp": true,
}

// ys6AlwaysCollectMetrics は、プレイヤー構造体のロード状態に関係なく収集を試みるメトリクスのキーです。
var ys6AlwaysCollectMetrics = []string{
	"boss_hp", "boss_max_hp", "mp_gauge",
	"hp", "hp_visual", "max_hp",
	"position_x", "position_y", "position_z", "on_ground",
}

// ys6PlayerStructMetrics は、プレイヤー構造体がロード済みの場合のみ収集するメトリクスのキーです。
var ys6PlayerStructMetrics = []string{
	"room_id", "level", "experience",
	"sword_livart_level", "sword_brillante_level", "sword_ericcil_level",
	"emel", "armor", "shield", "game_mode",
	"equipped_sword", "sword_livart_gauge", "sword_brillante_gauge", "sword_ericcil_gauge",
}

// Ys6Collector は、Ys VI: The Ark of Napishtim (YS6_WIN.exe) のゲームメトリクスを収集するCollectorの実装です。
// カプセル化の原則に従い、Ys6専用メトリクス収集という単一の目的を持ちます。
type Ys6Collector struct {
	resolver version.AddressResolver
	reader   memory.MemoryReader
	metrics  []Metric
}

// NewYs6Collector は、Ys6Collectorの新しいインスタンスを作成します。
// 関心の分離の原則に従い、アドレス解決とメモリ読み取りを外部に委譲します。
func NewYs6Collector(resolver version.AddressResolver, reader memory.MemoryReader) *Ys6Collector {
	return &Ys6Collector{
		resolver: resolver,
		reader:   reader,
		metrics:  make([]Metric, 0),
	}
}

// Collect は、Ys6のゲームメトリクスを収集します。
func (c *Ys6Collector) Collect() error {
	c.metrics = make([]Metric, 0, len(ys6AlwaysCollectMetrics)+len(ys6PlayerStructMetrics)+1)

	loaded := c.collectPlayerBase()

	for _, key := range ys6AlwaysCollectMetrics {
		c.collectMetric(key, ys6ZeroOnFailureMetrics[key])
	}

	if loaded {
		for _, key := range ys6PlayerStructMetrics {
			c.collectMetric(key, false)
		}
	}

	return nil
}

// GetMetrics は、収集済みのメトリクスを返します。
func (c *Ys6Collector) GetMetrics() []Metric {
	return c.metrics
}

// collectPlayerBase は、player_base の値を読み取り ys6_loaded メトリクスを出力します。
// アドレス解決または読み取りに失敗した場合は、未ロードとして扱います。
func (c *Ys6Collector) collectPlayerBase() bool {
	value, err := c.readInt32(playerBaseKey)
	loaded := err == nil && value != 0

	loadedValue := float64(0)
	if loaded {
		loadedValue = 1
	}

	c.metrics = append(c.metrics, Metric{
		Name:   ys6MetricName(playerBaseKey),
		Value:  loadedValue,
		Type:   Gauge,
		Labels: make(map[string]string),
	})

	return loaded
}

// playerBaseKey は、プレイヤー構造体ポインタを表すアドレス設定のキーです。
const playerBaseKey = "player_base"

// collectMetric は、指定キーのメトリクスを収集しmetricsに追加します。
// zeroOnFailure が true の場合、読み取り失敗時に0を出力します。falseの場合は欠損として出力しません。
func (c *Ys6Collector) collectMetric(key string, zeroOnFailure bool) {
	cfg, err := c.resolver.ResolveAddress(key)
	if err != nil {
		c.emitOnFailure(key, zeroOnFailure)
		return
	}

	address, err := c.resolveFinalAddress(cfg)
	if err != nil {
		c.emitOnFailure(key, zeroOnFailure)
		return
	}

	value, err := c.readByType(address, cfg.Type)
	if err != nil {
		c.emitOnFailure(key, zeroOnFailure)
		return
	}

	c.metrics = append(c.metrics, Metric{
		Name:   ys6MetricName(key),
		Value:  value,
		Type:   Gauge,
		Labels: make(map[string]string),
	})
}

// emitOnFailure は、zeroOnFailureがtrueの場合のみ値0のメトリクスを出力します。
func (c *Ys6Collector) emitOnFailure(key string, zeroOnFailure bool) {
	if !zeroOnFailure {
		return
	}
	c.metrics = append(c.metrics, Metric{
		Name:   ys6MetricName(key),
		Value:  0,
		Type:   Gauge,
		Labels: make(map[string]string),
	})
}

// readInt32 は、指定キーのアドレスを解決しint32値を読み取ります。
func (c *Ys6Collector) readInt32(key string) (int32, error) {
	cfg, err := c.resolver.ResolveAddress(key)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve address for %s: %w", key, err)
	}

	address, err := c.resolveFinalAddress(cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve final address for %s: %w", key, err)
	}

	value, err := c.reader.ReadInt32(address)
	if err != nil {
		return 0, fmt.Errorf("failed to read int32 for %s: %w", key, err)
	}

	return value, nil
}

// resolveFinalAddress は、AddressConfigから最終的な読み取りアドレスを解決します。
//
// BaseAdd が指定されている場合、Base と BaseAdd それぞれが指すアドレスから
// 32bit値を読み取り加算した値をベースポインタ P とします。
// Offsets[0] はデリファレンスなしで P に単純加算し、Offsets[1:] は
// 通常のポインタチェインとして解決します。
//
// BaseAdd が指定されていない場合は、従来どおり Base を起点に Offsets を
// ポインタチェインとして解決します(Offsetsが空の場合はBaseをそのまま返す)。
func (c *Ys6Collector) resolveFinalAddress(cfg *config.AddressConfig) (uintptr, error) {
	baseAddress, err := c.reader.ParseAndResolveAddress(cfg.Base)
	if err != nil {
		return 0, fmt.Errorf("failed to parse base address: %w", err)
	}

	offsets := cfg.Offsets
	current := baseAddress

	if cfg.BaseAdd != nil {
		baseAddAddress, err := c.reader.ParseAndResolveAddress(cfg.BaseAdd)
		if err != nil {
			return 0, fmt.Errorf("failed to parse base_add address: %w", err)
		}

		current, err = c.reader.ResolveSumBase(baseAddress, baseAddAddress)
		if err != nil {
			return 0, fmt.Errorf("failed to resolve summed base: %w", err)
		}

		if len(offsets) == 0 {
			return current, nil
		}

		current += uintptr(offsets[0])
		offsets = offsets[1:]
	}

	if len(offsets) == 0 {
		return current, nil
	}

	chainOffsets := make([]uintptr, len(offsets))
	for i, offset := range offsets {
		chainOffsets[i] = uintptr(offset)
	}

	return c.reader.ResolvePointerChain(current, chainOffsets)
}

// readByType は、指定されたアドレスから指定された型のデータを読み取ります。
func (c *Ys6Collector) readByType(address uintptr, dataType string) (float64, error) {
	switch dataType {
	case "int32":
		v, err := c.reader.ReadInt32(address)
		if err != nil {
			return 0, fmt.Errorf("failed to read int32: %w", err)
		}
		return float64(v), nil

	case "float32":
		v, err := c.reader.ReadFloat32(address)
		if err != nil {
			return 0, fmt.Errorf("failed to read float32: %w", err)
		}
		return float64(v), nil

	case "byte":
		v, err := c.reader.ReadByte(address)
		if err != nil {
			return 0, fmt.Errorf("failed to read byte: %w", err)
		}
		return float64(v), nil

	default:
		return 0, fmt.Errorf("unsupported data type: %s", dataType)
	}
}

// ys6MetricName は、内部メトリクス名をPrometheusフレンドリーな名前にマッピングします。
func ys6MetricName(key string) string {
	switch key {
	case "room_id":
		return "ys6_room_id"
	case "level":
		return "ys6_level"
	case "experience":
		return "ys6_exp"
	case "hp":
		return "ys6_hp"
	case "hp_visual":
		return "ys6_hp_visual"
	case "mp_gauge":
		return "ys6_mp_gauge"
	case "boss_hp":
		return "ys6_boss_hp"
	case "boss_max_hp":
		return "ys6_boss_max_hp"
	case "emel":
		return "ys6_emel"
	case "sword_livart_level":
		return "ys6_sword_livart_level"
	case "sword_brillante_level":
		return "ys6_sword_brillante_level"
	case "sword_ericcil_level":
		return "ys6_sword_ericcil_level"
	case "armor":
		return "ys6_armor"
	case "shield":
		return "ys6_shield"
	case "game_mode":
		return "ys6_game_mode"
	case "position_x":
		return "ys6_position_x"
	case "position_y":
		return "ys6_position_y"
	case "position_z":
		return "ys6_position_z"
	case "max_hp":
		return "ys6_max_hp"
	case "on_ground":
		return "ys6_on_ground"
	case "equipped_sword":
		return "ys6_equipped_sword"
	case "sword_livart_gauge":
		return "ys6_sword_livart_gauge"
	case "sword_brillante_gauge":
		return "ys6_sword_brillante_gauge"
	case "sword_ericcil_gauge":
		return "ys6_sword_ericcil_gauge"
	case playerBaseKey:
		return "ys6_loaded"
	default:
		return "ys6_" + key
	}
}
