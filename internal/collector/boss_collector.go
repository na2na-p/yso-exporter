package collector

import (
	"fmt"

	"github.com/na2na/yso-exporter/internal/memory"
	"github.com/na2na/yso-exporter/internal/version"
)

// コンパイル時にBossCollectorがCollectorインターフェースを実装していることを確認
var _ Collector = (*BossCollector)(nil)

// BossCollector は、ボス戦関連メトリクス（ボスHP、戦闘状態、ボスID）を収集するCollectorの実装です。
// カプセル化の原則に従い、ボス戦メトリクス収集という単一の目的を持ちます。
type BossCollector struct {
	resolver version.AddressResolver
	reader   memory.MemoryReader
	metrics  []Metric
}

// NewBossCollector は、BossCollectorの新しいインスタンスを作成します。
// 関心の分離の原則に従い、アドレス解決とメモリ読み取りを外部に委譲します。
func NewBossCollector(resolver version.AddressResolver, reader memory.MemoryReader) *BossCollector {
	return &BossCollector{
		resolver: resolver,
		reader:   reader,
		metrics:  make([]Metric, 0),
	}
}

// Collect は、ボス戦関連メトリクスを収集します。
// 収集対象: boss_hp, boss_max_hp, boss_id, in_boss_fight
func (bc *BossCollector) Collect() error {
	// メトリクスをクリア
	bc.metrics = make([]Metric, 0, 4)

	// 収集するボス戦メトリクスの定義
	metricNames := []string{"boss_hp", "boss_max_hp", "boss_id", "in_boss_fight"}

	for _, metricName := range metricNames {
		// アドレス設定を解決
		addressConfig, err := bc.resolver.ResolveAddress(metricName)
		if err != nil {
			return fmt.Errorf("failed to resolve address for %s: %w", metricName, err)
		}

		// Parse base address
		baseAddress, err := bc.reader.ParseAndResolveAddress(addressConfig.Base)
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
			finalAddress, err = bc.reader.ResolvePointerChain(baseAddress, offsets)
			if err != nil {
				return fmt.Errorf("failed to resolve pointer chain for %s: %w", metricName, err)
			}
		}

		// データ型に応じてメモリを読み取り、メトリクスを作成
		metric, err := bc.readMetric(metricName, finalAddress, addressConfig.Type)
		if err != nil {
			return fmt.Errorf("failed to read metric %s: %w", metricName, err)
		}

		bc.metrics = append(bc.metrics, metric)
	}

	return nil
}

// GetMetrics は、収集済みのメトリクスを返します。
func (bc *BossCollector) GetMetrics() []Metric {
	return bc.metrics
}

// readMetric は、指定されたアドレスから指定された型のデータを読み取り、Metricとして返します。
func (bc *BossCollector) readMetric(name string, address uintptr, dataType string) (Metric, error) {
	var value float64
	var labels map[string]string = make(map[string]string)
	var metricType MetricType = Gauge

	switch dataType {
	case "byte":
		byteVal, err := bc.reader.ReadByte(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read byte: %w", err)
		}
		value = float64(byteVal)

		// ボスIDまたは戦闘状態の場合、ラベルを追加
		if name == "boss_id" {
			bossName := bc.getBossName(byteVal)
			labels["boss"] = bossName
		} else if name == "in_boss_fight" {
			labels["status"] = bc.getBattleStatus(byteVal)
		}

	case "int32":
		intVal, err := bc.reader.ReadInt32(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read int32: %w", err)
		}
		value = float64(intVal)

	case "float32":
		floatVal, err := bc.reader.ReadFloat32(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read float32: %w", err)
		}
		value = float64(floatVal)

	default:
		return Metric{}, fmt.Errorf("unsupported data type: %s", dataType)
	}

	// ボス戦タイプに応じてラベルを追加
	labels["metric_type"] = bc.getMetricType(name)

	return Metric{
		Name:   name,
		Value:  value,
		Type:   metricType,
		Labels: labels,
	}, nil
}

// getBossName は、ボスIDから文字列名を返します。
func (bc *BossCollector) getBossName(bossID byte) string {
	// Ys Origin のボスID定義（例）
	switch bossID {
	case 0:
		return "none"
	case 1:
		return "gyalva"
	case 2:
		return "epona"
	case 3:
		return "kishgal"
	case 4:
		return "zava"
	case 5:
		return "omulgun"
	case 6:
		return "majunun"
	case 7:
		return "rud_um"
	case 8:
		return "dalles"
	case 9:
		return "darm"
	default:
		return "unknown"
	}
}

// getBattleStatus は、戦闘状態から文字列を返します。
func (bc *BossCollector) getBattleStatus(status byte) string {
	switch status {
	case 0:
		return "normal"
	case 1:
		return "boss_fight"
	default:
		return "unknown"
	}
}

// getMetricType は、ボス戦メトリクス名からタイプ分類を返します。
func (bc *BossCollector) getMetricType(metricName string) string {
	switch metricName {
	case "boss_hp", "boss_max_hp":
		return "health"
	case "boss_id":
		return "identifier"
	case "in_boss_fight":
		return "state"
	default:
		return "unknown"
	}
}