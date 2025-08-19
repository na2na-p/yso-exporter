package collector

import (
	"fmt"

	"github.com/na2na/yso-exporter/internal/memory"
	"github.com/na2na/yso-exporter/internal/version"
)

// コンパイル時にPositionCollectorがCollectorインターフェースを実装していることを確認
var _ Collector = (*PositionCollector)(nil)

// PositionCollector は、位置情報メトリクス（X、Y座標、現在のマップ）を収集するCollectorの実装です。
// カプセル化の原則に従い、位置情報収集という単一の目的を持ちます。
type PositionCollector struct {
	resolver version.AddressResolver
	reader   memory.MemoryReader
	metrics  []Metric
}

// NewPositionCollector は、PositionCollectorの新しいインスタンスを作成します。
// 関心の分離の原則に従い、アドレス解決とメモリ読み取りを外部に委譲します。
func NewPositionCollector(resolver version.AddressResolver, reader memory.MemoryReader) *PositionCollector {
	return &PositionCollector{
		resolver: resolver,
		reader:   reader,
		metrics:  make([]Metric, 0),
	}
}

// Collect は、位置情報メトリクスを収集します。
// 収集対象: position_x, position_y, current_map
func (pc *PositionCollector) Collect() error {
	// メトリクスをクリア
	pc.metrics = make([]Metric, 0, 3)

	// 収集する位置メトリクスの定義
	metricNames := []string{"position_x", "position_y", "current_map"}

	for _, metricName := range metricNames {
		// アドレス設定を解決
		addressConfig, err := pc.resolver.ResolveAddress(metricName)
		if err != nil {
			return fmt.Errorf("failed to resolve address for %s: %w", metricName, err)
		}

		// Parse base address
		baseAddress, err := pc.reader.ParseAndResolveAddress(addressConfig.Base)
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
			finalAddress, err = pc.reader.ResolvePointerChain(baseAddress, offsets)
			if err != nil {
				return fmt.Errorf("failed to resolve pointer chain for %s: %w", metricName, err)
			}
		}

		// データ型に応じてメモリを読み取り、メトリクスを作成
		metric, err := pc.readMetric(metricName, finalAddress, addressConfig.Type)
		if err != nil {
			return fmt.Errorf("failed to read metric %s: %w", metricName, err)
		}

		pc.metrics = append(pc.metrics, metric)
	}

	return nil
}

// GetMetrics は、収集済みのメトリクスを返します。
func (pc *PositionCollector) GetMetrics() []Metric {
	return pc.metrics
}

// readMetric は、指定されたアドレスから指定された型のデータを読み取り、Metricとして返します。
func (pc *PositionCollector) readMetric(name string, address uintptr, dataType string) (Metric, error) {
	var value float64
	var labels map[string]string = make(map[string]string)

	switch dataType {
	case "byte":
		byteVal, err := pc.reader.ReadByte(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read byte: %w", err)
		}
		value = float64(byteVal)

		// マップIDの場合、ラベルを追加
		if name == "current_map" {
			mapName := pc.getMapName(byteVal)
			labels["map"] = mapName
		}

	case "int32":
		intVal, err := pc.reader.ReadInt32(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read int32: %w", err)
		}
		value = float64(intVal)

	case "float32":
		floatVal, err := pc.reader.ReadFloat32(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read float32: %w", err)
		}
		value = float64(floatVal)

	default:
		return Metric{}, fmt.Errorf("unsupported data type: %s", dataType)
	}

	return Metric{
		Name:   name,
		Value:  value,
		Type:   Gauge, // 位置メトリクスは全てGaugeとして扱う
		Labels: labels,
	}, nil
}

// getMapName は、マップIDから文字列名を返します。
func (pc *PositionCollector) getMapName(mapID byte) string {
	// Ys Origin のマップID定義（例）
	switch mapID {
	case 0:
		return "darm_tower_1f"
	case 1:
		return "darm_tower_2f"
	case 2:
		return "darm_tower_3f"
	case 3:
		return "darm_tower_4f"
	case 4:
		return "darm_tower_5f"
	case 5:
		return "darm_tower_6f"
	case 10:
		return "silent_sands"
	case 11:
		return "burning_cave"
	case 12:
		return "submerged_city"
	default:
		return "unknown"
	}
}