package collector

import (
	"fmt"

	"github.com/na2na/yso-exporter/internal/memory"
	"github.com/na2na/yso-exporter/internal/version"
)

// コンパイル時にCombatCollectorがCollectorインターフェースを実装していることを確認
var _ Collector = (*CombatCollector)(nil)

// CombatCollector は、戦闘関連メトリクス（HP、MP、攻撃力、防御力）を収集するCollectorの実装です。
// カプセル化の原則に従い、戦闘メトリクス収集という単一の目的を持ちます。
type CombatCollector struct {
	resolver version.AddressResolver
	reader   memory.MemoryReader
	metrics  []Metric
}

// NewCombatCollector は、CombatCollectorの新しいインスタンスを作成します。
// 関心の分離の原則に従い、アドレス解決とメモリ読み取りを外部に委譲します。
func NewCombatCollector(resolver version.AddressResolver, reader memory.MemoryReader) *CombatCollector {
	return &CombatCollector{
		resolver: resolver,
		reader:   reader,
		metrics:  make([]Metric, 0),
	}
}

// Collect は、戦闘関連メトリクスを収集します。
// 収集対象: current_hp, current_mp, max_hp, max_mp, attack, defense
func (cc *CombatCollector) Collect() error {
	// メトリクスをクリア
	cc.metrics = make([]Metric, 0, 6)

	// 収集する戦闘メトリクスの定義
	metricNames := []string{"current_hp", "current_mp", "max_hp", "max_mp", "attack", "defense"}

	for _, metricName := range metricNames {
		// アドレス設定を解決
		addressConfig, err := cc.resolver.ResolveAddress(metricName)
		if err != nil {
			return fmt.Errorf("failed to resolve address for %s: %w", metricName, err)
		}

		// Parse base address
		baseAddress, err := cc.reader.ParseAndResolveAddress(addressConfig.Base)
		if err != nil {
			return fmt.Errorf("failed to parse base address for %s: %w", metricName, err)
		}

		// Resolve pointer chain if needed
		finalAddress := baseAddress
		if len(addressConfig.Offsets) > 0 {
			offsets := make([]uintptr, len(addressConfig.Offsets))
			for i, offset := range addressConfig.Offsets {
				offsets[i] = uintptr(offset)
			}
			finalAddress, err = cc.reader.ResolvePointerChain(baseAddress, offsets)
			if err != nil {
				return fmt.Errorf("failed to resolve pointer chain for %s: %w", metricName, err)
			}
		}

		// データ型に応じてメモリを読み取り、メトリクスを作成
		metric, err := cc.readMetric(metricName, finalAddress, addressConfig.Type)
		if err != nil {
			return fmt.Errorf("failed to read metric %s: %w", metricName, err)
		}

		cc.metrics = append(cc.metrics, metric)
	}

	return nil
}

// GetMetrics は、収集済みのメトリクスを返します。
func (cc *CombatCollector) GetMetrics() []Metric {
	return cc.metrics
}

// readMetric は、指定されたアドレスから指定された型のデータを読み取り、Metricとして返します。
func (cc *CombatCollector) readMetric(name string, address uintptr, dataType string) (Metric, error) {
	var value float64
	var labels map[string]string = make(map[string]string)
	var metricType MetricType = Gauge

	switch dataType {
	case "byte":
		byteVal, err := cc.reader.ReadByte(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read byte: %w", err)
		}
		value = float64(byteVal)

	case "int16":
		// int16は現在サポートされていないため、int32として読み取り
		intVal, err := cc.reader.ReadInt32(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read int32 (as int16): %w", err)
		}
		value = float64(intVal)

	case "int32":
		intVal, err := cc.reader.ReadInt32(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read int32: %w", err)
		}
		value = float64(intVal)

	case "float32":
		floatVal, err := cc.reader.ReadFloat32(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read float32: %w", err)
		}
		value = float64(floatVal)

	default:
		return Metric{}, fmt.Errorf("unsupported data type: %s", dataType)
	}

	// 戦闘タイプに応じてラベルを追加
	labels["stat_type"] = cc.getStatType(name)

	return Metric{
		Name:   name,
		Value:  value,
		Type:   metricType,
		Labels: labels,
	}, nil
}

// getStatType は、戦闘メトリクス名からタイプ分類を返します。
func (cc *CombatCollector) getStatType(metricName string) string {
	switch metricName {
	case "current_hp", "max_hp":
		return "health"
	case "current_mp", "max_mp":
		return "mana"
	case "attack":
		return "offense"
	case "defense":
		return "defense"
	default:
		return "unknown"
	}
}