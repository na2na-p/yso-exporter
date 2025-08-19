package collector

import (
	"fmt"

	"github.com/na2na/yso-exporter/internal/memory"
	"github.com/na2na/yso-exporter/internal/version"
)

// コンパイル時にFlagCollectorがCollectorインターフェースを実装していることを確認
var _ Collector = (*FlagCollector)(nil)

// FlagCollector は、進行フラグとテレポートフラグを収集するCollectorの実装です。
// カプセル化の原則に従い、フラグ収集という単一の目的を持ちます。
type FlagCollector struct {
	resolver version.AddressResolver
	reader   memory.MemoryReader
	metrics  []Metric
}

// NewFlagCollector は、FlagCollectorの新しいインスタンスを作成します。
// 関心の分離の原則に従い、アドレス解決とメモリ読み取りを外部に委譲します。
func NewFlagCollector(resolver version.AddressResolver, reader memory.MemoryReader) *FlagCollector {
	return &FlagCollector{
		resolver: resolver,
		reader:   reader,
		metrics:  make([]Metric, 0),
	}
}

// Collect は、フラグメトリクスを収集します。
// 収集対象: story_flags, teleport_flags, event_flags, item_flags
func (fc *FlagCollector) Collect() error {
	// メトリクスをクリア
	fc.metrics = make([]Metric, 0, 10) // フラグは複数の値を持つ可能性があるため多めに確保

	// 収集するフラグメトリクスの定義
	flagTypes := []string{"story_flags", "teleport_flags", "event_flags", "item_flags"}

	for _, flagType := range flagTypes {
		// アドレス設定を解決
		addressConfig, err := fc.resolver.ResolveAddress(flagType)
		if err != nil {
			return fmt.Errorf("failed to resolve address for %s: %w", flagType, err)
		}

		// Parse base address
		baseAddress, err := fc.reader.ParseAndResolveAddress(addressConfig.Base)
		if err != nil {
			return fmt.Errorf("failed to parse base address for %s: %w", flagType, err)
		}

		// Resolve pointer chain if needed
		finalAddress := baseAddress
		if len(addressConfig.Offsets) > 0 {
			offsets := make([]uintptr, len(addressConfig.Offsets))
			for i, offset := range addressConfig.Offsets {
				offsets[i] = uintptr(offset)
			}
			finalAddress, err = fc.reader.ResolvePointerChain(baseAddress, offsets)
			if err != nil {
				return fmt.Errorf("failed to resolve pointer chain for %s: %w", flagType, err)
			}
		}

		// フラグタイプに応じてメトリクスを収集
		err = fc.collectFlagMetrics(flagType, finalAddress, addressConfig.Type)
		if err != nil {
			return fmt.Errorf("failed to collect flag metrics for %s: %w", flagType, err)
		}
	}

	return nil
}

// GetMetrics は、収集済みのメトリクスを返します。
func (fc *FlagCollector) GetMetrics() []Metric {
	return fc.metrics
}

// collectFlagMetrics は、指定されたアドレスからフラグデータを読み取り、個別のメトリクスとして追加します。
func (fc *FlagCollector) collectFlagMetrics(flagType string, address uintptr, dataType string) error {
	switch dataType {
	case "byte_array":
		// バイト配列として複数のフラグバイトを読み取り
		return fc.collectByteArrayFlags(flagType, address)
	case "int32":
		// 単一のint32値として読み取り
		return fc.collectSingleFlag(flagType, address, "int32")
	case "byte":
		// 単一のbyte値として読み取り
		return fc.collectSingleFlag(flagType, address, "byte")
	default:
		return fmt.Errorf("unsupported flag data type: %s", dataType)
	}
}

// collectByteArrayFlags は、バイト配列からフラグを収集します。
func (fc *FlagCollector) collectByteArrayFlags(flagType string, baseAddress uintptr) error {
	// フラグタイプに応じた配列サイズを決定
	arraySize := fc.getFlagArraySize(flagType)
	
	for i := 0; i < arraySize; i++ {
		address := baseAddress + uintptr(i)
		
		byteVal, err := fc.reader.ReadByte(address)
		if err != nil {
			return fmt.Errorf("failed to read flag byte at offset %d: %w", i, err)
		}

		// 各ビットをフラグとして解析
		for bit := 0; bit < 8; bit++ {
			flagValue := (byteVal >> bit) & 1
			flagIndex := i*8 + bit
			
			flagName := fmt.Sprintf("%s_%d", flagType, flagIndex)
			labels := map[string]string{
				"flag_type":  flagType,
				"flag_index": fmt.Sprintf("%d", flagIndex),
				"flag_name":  fc.getFlagName(flagType, flagIndex),
			}

			metric := Metric{
				Name:   flagName,
				Value:  float64(flagValue),
				Type:   Gauge,
				Labels: labels,
			}

			fc.metrics = append(fc.metrics, metric)
		}
	}

	return nil
}

// collectSingleFlag は、単一のフラグ値を収集します。
func (fc *FlagCollector) collectSingleFlag(flagType string, address uintptr, dataType string) error {
	var value float64

	switch dataType {
	case "byte":
		byteVal, readErr := fc.reader.ReadByte(address)
		if readErr != nil {
			return fmt.Errorf("failed to read byte flag: %w", readErr)
		}
		value = float64(byteVal)
	case "int32":
		intVal, readErr := fc.reader.ReadInt32(address)
		if readErr != nil {
			return fmt.Errorf("failed to read int32 flag: %w", readErr)
		}
		value = float64(intVal)
	default:
		return fmt.Errorf("unsupported single flag type: %s", dataType)
	}

	labels := map[string]string{
		"flag_type": flagType,
	}

	metric := Metric{
		Name:   flagType,
		Value:  value,
		Type:   Gauge,
		Labels: labels,
	}

	fc.metrics = append(fc.metrics, metric)
	return nil
}

// getFlagArraySize は、フラグタイプに応じた配列サイズを返します。
func (fc *FlagCollector) getFlagArraySize(flagType string) int {
	switch flagType {
	case "story_flags":
		return 10 // 80個のストーリーフラグ
	case "teleport_flags":
		return 5  // 40個のテレポートフラグ
	case "event_flags":
		return 15 // 120個のイベントフラグ
	case "item_flags":
		return 20 // 160個のアイテムフラグ
	default:
		return 4 // デフォルトで32個のフラグ
	}
}

// getFlagName は、フラグタイプとインデックスから意味のある名前を返します。
func (fc *FlagCollector) getFlagName(flagType string, index int) string {
	// 実際のゲームデータに基づいてフラグ名を定義
	switch flagType {
	case "story_flags":
		return fc.getStoryFlagName(index)
	case "teleport_flags":
		return fc.getTeleportFlagName(index)
	case "event_flags":
		return fc.getEventFlagName(index)
	case "item_flags":
		return fc.getItemFlagName(index)
	default:
		return fmt.Sprintf("flag_%d", index)
	}
}

// getStoryFlagName は、ストーリーフラグの名前を返します。
func (fc *FlagCollector) getStoryFlagName(index int) string {
	storyFlags := map[int]string{
		0:  "game_start",
		1:  "met_yunica",
		2:  "met_hugo",
		3:  "tower_entrance_unlocked",
		5:  "first_boss_defeated",
		10: "second_floor_unlocked",
		15: "third_floor_unlocked",
		20: "final_boss_unlocked",
	}
	
	if name, exists := storyFlags[index]; exists {
		return name
	}
	return fmt.Sprintf("story_%d", index)
}

// getTeleportFlagName は、テレポートフラグの名前を返します。
func (fc *FlagCollector) getTeleportFlagName(index int) string {
	teleportFlags := map[int]string{
		0: "tower_1f_teleport",
		1: "tower_2f_teleport",
		2: "tower_3f_teleport",
		3: "tower_4f_teleport",
		4: "tower_5f_teleport",
		5: "silent_sands_teleport",
		6: "burning_cave_teleport",
		7: "submerged_city_teleport",
	}
	
	if name, exists := teleportFlags[index]; exists {
		return name
	}
	return fmt.Sprintf("teleport_%d", index)
}

// getEventFlagName は、イベントフラグの名前を返します。
func (fc *FlagCollector) getEventFlagName(index int) string {
	eventFlags := map[int]string{
		0:  "intro_watched",
		1:  "tutorial_completed",
		5:  "first_cutscene_watched",
		10: "boss_intro_watched",
		15: "ending_unlocked",
	}
	
	if name, exists := eventFlags[index]; exists {
		return name
	}
	return fmt.Sprintf("event_%d", index)
}

// getItemFlagName は、アイテムフラグの名前を返します。
func (fc *FlagCollector) getItemFlagName(index int) string {
	itemFlags := map[int]string{
		0:  "bronze_sword_obtained",
		1:  "silver_sword_obtained",
		2:  "wind_sword_obtained",
		10: "bronze_armor_obtained",
		11: "silver_armor_obtained",
		20: "fire_magic_obtained",
		21: "ice_magic_obtained",
		22: "thunder_magic_obtained",
	}
	
	if name, exists := itemFlags[index]; exists {
		return name
	}
	return fmt.Sprintf("item_%d", index)
}