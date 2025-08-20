package collector

import (
	"fmt"
	"strings"

	"github.com/na2na/yso-exporter/internal/memory"
	"github.com/na2na/yso-exporter/internal/version"
)

// コンパイル時にBasicCollectorがCollectorインターフェースを実装していることを確認
var _ Collector = (*BasicCollector)(nil)

// BasicCollector は、基本的なゲームメトリクス（キャラクター、レベル、HP等）を収集するCollectorの実装です。
// カプセル化の原則に従い、基本メトリクス収集という単一の目的を持ちます。
type BasicCollector struct {
	resolver version.AddressResolver
	reader   memory.MemoryReader
	metrics  []Metric
}

// NewBasicCollector は、BasicCollectorの新しいインスタンスを作成します。
// 関心の分離の原則に従い、アドレス解決とメモリ読み取りを外部に委譲します。
func NewBasicCollector(resolver version.AddressResolver, reader memory.MemoryReader) *BasicCollector {
	return &BasicCollector{
		resolver: resolver,
		reader:   reader,
		metrics:  make([]Metric, 0),
	}
}

// Collect は、基本的なゲームメトリクスを収集します。
// 収集対象: character, level, exp, hp, mp
func (bc *BasicCollector) Collect() error {
	// メトリクスをクリア
	bc.metrics = make([]Metric, 0, 12)

	// 収集する基本メトリクスの定義
	// 注: コンフィグファイルのキー名と一致させる必要がある
	metricNames := []string{"character", "level", "experience", "hp", "mp", "ground_pointer", "boost_gauge",
		"current_equip_skill", "floor_boss_hp_gauge", "boss_hp_some", "velg_pict_hp",
		"str_boost", "def_boost", "mp_boost", "hit_count",
		"str_boost_visual", "def_boost_visual", "mp_boost_visual", "hit_count_visual"}

	// プロセス接続チェックのために最初のメトリクスで試す
	hasProcessError := false

	for i, metricName := range metricNames {
		// アドレス設定を解決
		addressConfig, err := bc.resolver.ResolveAddress(metricName)
		if err != nil {
			// アドレスが見つからない場合はスキップ（ボスHPなど、状況によって存在しないメトリクス対応）
			// fmt.Printf("Failed to resolve address for %s: %v\n", metricName, err)
			continue
		}

		// ベースアドレスを解析
		baseAddress, err := bc.reader.ParseAndResolveAddress(addressConfig.Base)
		if err != nil {
			// プロセスが見つからないエラーの場合
			if i == 0 && (strings.Contains(err.Error(), "not found in process") || strings.Contains(err.Error(), "module")) {
				// 最初のメトリクスでプロセスエラーなら、全体をスキップ
				return fmt.Errorf("failed to parse base address for %s: %w", metricName, err)
			}
			// その他のメトリクスでのエラーは個別に処理
			if !hasProcessError {
				return fmt.Errorf("failed to parse base address for %s: %w", metricName, err)
			}
			continue
		}

		// ポインタチェインを解決してアドレスを取得
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
			// 読み取りエラーの場合は0で出力（ボスHPなど、ボスがいない時は読めない）
			if metricName == "velg_pict_hp" || metricName == "boss_hp_some" ||
				metricName == "str_boost" || metricName == "def_boost" ||
				metricName == "mp_boost" || metricName == "hit_count" ||
				metricName == "str_boost_visual" || metricName == "def_boost_visual" ||
				metricName == "mp_boost_visual" || metricName == "hit_count_visual" {
				// ボス系・ブースト系メトリクスは0で出力
				metric = Metric{
					Name:   bc.mapMetricName(metricName),
					Value:  0,
					Type:   Gauge,
					Labels: make(map[string]string),
				}
				bc.metrics = append(bc.metrics, metric)
			}
			// current_equip_skillは読み取りエラー時はメトリクスを出力しない（欠損データとする）
			continue
		}

		bc.metrics = append(bc.metrics, metric)
	}

	return nil
}

// GetMetrics は、収集済みのメトリクスを返します。
func (bc *BasicCollector) GetMetrics() []Metric {
	return bc.metrics
}

// readMetric は、指定されたアドレスから指定された型のデータを読み取り、Metricとして返します。
func (bc *BasicCollector) readMetric(name string, address uintptr, dataType string) (Metric, error) {
	var value float64
	var labels map[string]string = make(map[string]string)

	switch dataType {
	case "byte":
		byteVal, err := bc.reader.ReadByte(address)
		if err != nil {
			return Metric{}, fmt.Errorf("failed to read byte: %w", err)
		}
		value = float64(byteVal)

		// キャラクターの場合もラベルを追加しない（値だけで判別）
		// if name == "character" {
		// 	characterName := bc.getCharacterName(byteVal)
		// 	labels["character"] = characterName
		// }

		// 装備スキルの場合はラベルを追加しない（値だけで判別）
		// if name == "current_equip_skill" {
		// 	skillName := bc.getSkillName(byteVal)
		// 	labels["skill"] = skillName
		// }

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

	// メトリクス名をPrometheusフレンドリーな名前にマッピング
	prometheusName := bc.mapMetricName(name)

	return Metric{
		Name:   prometheusName,
		Value:  value,
		Type:   Gauge, // 基本メトリクスは全てGaugeとして扱う
		Labels: labels,
	}, nil
}

// mapMetricName は、内部メトリクス名をPrometheusフレンドリーな名前にマッピングします。
func (bc *BasicCollector) mapMetricName(name string) string {
	switch name {
	case "character":
		return "yso_character"
	case "level":
		return "yso_level"
	case "experience":
		return "yso_exp"
	case "hp":
		return "yso_hp"
	case "mp":
		return "yso_mp"
	case "ground_pointer":
		return "yso_on_ground"
	case "boost_gauge":
		return "yso_boost"
	case "floor_boss_hp_gauge":
		return "yso_boss_hp_floor"
	case "boss_hp_some":
		return "yso_boss_hp"
	case "velg_pict_hp":
		return "yso_boss_hp_velg_pict"
	case "str_boost":
		return "yso_str_boost"
	case "def_boost":
		return "yso_def_boost"
	case "mp_boost":
		return "yso_mp_boost"
	case "hit_count":
		return "yso_exp_rate"
	case "hit_count_visual":
		return "yso_exp_rate_time"
	case "str_boost_visual":
		return "yso_str_boost_time"
	case "def_boost_visual":
		return "yso_def_boost_time"
	case "mp_boost_visual":
		return "yso_mp_boost_time"
	case "current_equip_skill":
		return "yso_equipped_skill"
	default:
		// デフォルトはプレフィックスを付けるだけ
		return "yso_" + name
	}
}

// getCharacterName は、キャラクターIDから文字列名を返します。
func (bc *BasicCollector) getCharacterName(characterID byte) string {
	switch characterID {
	case 0:
		return "adol"
	case 1:
		return "yunica"
	case 2:
		return "hugo"
	case 3:
		return "toal"
	default:
		return "unknown"
	}
}

// getSkillName は、スキルIDから文字列名を返します。
func (bc *BasicCollector) getSkillName(skillID byte) string {
	switch skillID {
	case 116:
		return "wind"
	case 117:
		return "thunder"
	case 118:
		return "fire"
	case 255:
		return "none"
	default:
		// その他の値は数値として表示（デバッグ用）
		return fmt.Sprintf("skill_%d", skillID)
	}
}
