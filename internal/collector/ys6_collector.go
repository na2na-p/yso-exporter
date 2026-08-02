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

// ys6MaxRetainedFailures は、直前値を保持したまま出力し続ける連続失敗サイクル数の上限です。
// ポーリング間隔300msの下で2サイクル(約0.6秒)を許容上限とすることで、
// ゲーム終了後に古い値を出し続ける時間をこの範囲に抑えます。
const ys6MaxRetainedFailures = 2

// ys6MetricRetention は、1メトリクスキー分の「直前の成功値」と「連続失敗サイクル数」を保持します。
type ys6MetricRetention struct {
	hasValue bool
	value    float64
	failures int
}

// Ys6Collector は、Ys VI: The Ark of Napishtim (YS6_WIN.exe) のゲームメトリクスを収集するCollectorの実装です。
// カプセル化の原則に従い、Ys6専用メトリクス収集という単一の目的を持ちます。
type Ys6Collector struct {
	resolver version.AddressResolver
	reader   memory.MemoryReader
	metrics  []Metric

	// retentionStates は、zeroOnFailureでないメトリクスの直前値保持状態をキーごとに保持します。
	// Collectの呼び出し元はSnapshotCollector経由のPoller1本に直列化されるため、
	// ここに排他制御は設けていません。
	retentionStates map[string]*ys6MetricRetention

	// playerBaseFailures と playerBaseLoaded は、player_base自体の読み取りが一時的に失敗した際に
	// 直前のロード状態を維持するための状態です。
	playerBaseFailures int
	playerBaseLoaded   bool
}

// NewYs6Collector は、Ys6Collectorの新しいインスタンスを作成します。
// 関心の分離の原則に従い、アドレス解決とメモリ読み取りを外部に委譲します。
func NewYs6Collector(resolver version.AddressResolver, reader memory.MemoryReader) *Ys6Collector {
	return &Ys6Collector{
		resolver:        resolver,
		reader:          reader,
		metrics:         make([]Metric, 0),
		retentionStates: make(map[string]*ys6MetricRetention),
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

// collectPlayerBase は、player_base(BaseとBaseAddの合算値)を解決し ys6_loaded メトリクスを出力します。
//
// Base単体([177A44]相当)はプレイ中でも0〜0xF0の間で頻繁に変動し、0はプレイ中の正常値でもあるため、
// 実体ポインタを持つBaseAdd([177A48]相当)との合算値で判定します。合算値自体がプレイヤー構造体の
// 実アドレスであるため、アドレスとしてデリファレンスせず非ゼロかどうかのみで判定します。
//
// 解決成功時は合算値そのものに基づき判定します(合算値0は本物のアンロードとして、
// 全メトリクスの保持値・失敗カウンタをクリアします)。
// 解決失敗時は一時的な失敗とみなし、連続失敗が ys6MaxRetainedFailures 以内なら
// 直前のロード状態を維持します(直前がロード中ならプレイヤー系メトリクスの収集も通常どおり試行します)。
// 連続失敗がそれを超えたら未ロード扱いとし、保持値をクリアします。
func (c *Ys6Collector) collectPlayerBase() bool {
	sum, err := c.readPlayerBaseSum()

	var loaded bool
	if err == nil {
		loaded = sum != 0
		c.playerBaseFailures = 0
		if !loaded {
			c.clearRetention()
		}
	} else {
		c.playerBaseFailures++
		if c.playerBaseFailures <= ys6MaxRetainedFailures {
			loaded = c.playerBaseLoaded
		} else {
			loaded = false
			c.clearRetention()
		}
	}
	c.playerBaseLoaded = loaded

	loadedValue := float64(0)
	if loaded {
		loadedValue = 1
	}
	c.appendMetric(playerBaseKey, loadedValue)

	return loaded
}

// clearRetention は、全メトリクスの直前値保持状態を破棄します。
// プレイヤー構造体が本物のアンロード(または諦め)に至った際、次のロード後に
// 古いサイクルの値が漏れ出さないようにするために呼び出します。
func (c *Ys6Collector) clearRetention() {
	c.retentionStates = make(map[string]*ys6MetricRetention)
}

// playerBaseKey は、プレイヤー構造体ポインタを表すアドレス設定のキーです。
const playerBaseKey = "player_base"

// readPlayerBaseSum は、player_base の Base と BaseAdd がそれぞれ指すアドレスから
// 32bit値を読み取り合算した値を返します。合算値自体がプレイヤー構造体の実アドレスであり、
// 非ゼロかどうかでロード判定に用いるため、この値をさらにアドレスとして読み取ってはいけません。
func (c *Ys6Collector) readPlayerBaseSum() (uintptr, error) {
	cfg, err := c.resolver.ResolveAddress(playerBaseKey)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve address for %s: %w", playerBaseKey, err)
	}

	return c.resolveSumAddress(cfg)
}

// collectMetric は、指定キーのメトリクスを収集しmetricsに追加します。
// zeroOnFailure が true の場合、読み取り失敗時に0を出力します。
// falseの場合は、直前の成功値を ys6MaxRetainedFailures サイクルまで保持して出力し、
// それを超えて連続失敗したら欠損として出力しません。
func (c *Ys6Collector) collectMetric(key string, zeroOnFailure bool) {
	value, err := c.readMetricValue(key)

	if zeroOnFailure {
		if err != nil {
			c.appendMetric(key, 0)
			return
		}
		c.appendMetric(key, value)
		return
	}

	c.collectWithRetention(key, value, err)
}

// collectWithRetention は、直前値保持を伴うメトリクス収集を行います。
// 読み取り成功時は保持値を更新し失敗カウンタをリセットします。
// 読み取り失敗時は失敗カウンタを増やし、保持値があり許容範囲内であればその値を出力します。
func (c *Ys6Collector) collectWithRetention(key string, value float64, err error) {
	state, ok := c.retentionStates[key]
	if !ok {
		state = &ys6MetricRetention{}
		c.retentionStates[key] = state
	}

	if err == nil {
		state.hasValue = true
		state.value = value
		state.failures = 0
		c.appendMetric(key, value)
		return
	}

	state.failures++
	if state.hasValue && state.failures <= ys6MaxRetainedFailures {
		c.appendMetric(key, state.value)
	}
}

// readMetricValue は、指定キーのアドレスを解決し値を読み取ります。
// アドレス解決・最終アドレス解決・値の読み取りのいずれかに失敗した場合はエラーを返します。
func (c *Ys6Collector) readMetricValue(key string) (float64, error) {
	cfg, err := c.resolver.ResolveAddress(key)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve address for %s: %w", key, err)
	}

	address, err := c.resolveFinalAddress(cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve final address for %s: %w", key, err)
	}

	value, err := c.readByType(address, cfg.Type)
	if err != nil {
		return 0, fmt.Errorf("failed to read value for %s: %w", key, err)
	}

	return value, nil
}

// appendMetric は、指定キー・値のメトリクスをmetricsへ追加します。
func (c *Ys6Collector) appendMetric(key string, value float64) {
	c.metrics = append(c.metrics, Metric{
		Name:   ys6MetricName(key),
		Value:  value,
		Type:   Gauge,
		Labels: make(map[string]string),
	})
}

// resolveSumAddress は、cfg.Base と(指定されていれば)cfg.BaseAdd がそれぞれ指すアドレスから
// 32bit値を読み取り合算した値を返します。BaseAdd 未指定時は Base の解決値のみを返します。
// resolveFinalAddress のBaseAdd合算処理と player_base のゲーティング判定の両方が
// 同じ「Base+BaseAddの合算値を得る」という関心を持つため、ここに集約しています。
func (c *Ys6Collector) resolveSumAddress(cfg *config.AddressConfig) (uintptr, error) {
	baseAddress, err := c.reader.ParseAndResolveAddress(cfg.Base)
	if err != nil {
		return 0, fmt.Errorf("failed to parse base address: %w", err)
	}

	if cfg.BaseAdd == nil {
		return baseAddress, nil
	}

	baseAddAddress, err := c.reader.ParseAndResolveAddress(cfg.BaseAdd)
	if err != nil {
		return 0, fmt.Errorf("failed to parse base_add address: %w", err)
	}

	sum, err := c.reader.ResolveSumBase(baseAddress, baseAddAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve summed base: %w", err)
	}

	return sum, nil
}

// resolveFinalAddress は、AddressConfigから最終的な読み取りアドレスを解決します。
//
// BaseAdd が指定されている場合、合算値をベースポインタ P とし、Offsets[0] は
// デリファレンスなしで P に単純加算、Offsets[1:] は通常のポインタチェインとして解決します。
//
// BaseAdd が指定されていない場合は、従来どおり Base を起点に Offsets を
// ポインタチェインとして解決します(Offsetsが空の場合はBaseをそのまま返す)。
func (c *Ys6Collector) resolveFinalAddress(cfg *config.AddressConfig) (uintptr, error) {
	current, err := c.resolveSumAddress(cfg)
	if err != nil {
		return 0, err
	}

	offsets := cfg.Offsets

	if cfg.BaseAdd != nil {
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
