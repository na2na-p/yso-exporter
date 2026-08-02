package config

import "sort"

// BuiltinVersionConfigs contains all built-in version configurations
// These are the only supported configurations (no external YAML files)
var BuiltinVersionConfigs = map[string]*VersionConfig{
	"jp_1.1.0.3": {
		Metadata: struct {
			Name        string `yaml:"name"`
			DisplayName string `yaml:"display_name"`
			Description string `yaml:"description"`
			ProcessName string `yaml:"process_name"`
			Game        string `yaml:"game"`
		}{
			Name:        "jp_1.1.0.3",
			DisplayName: "Japanese v1.1.0.3",
			Description: "Ys Origin Japanese version 1.1.0.3 (日本語版)",
			ProcessName: "YSO_WIN.exe",
			Game:        "yso",
		},
		Addresses: map[string]AddressConfig{
			"character": {
				Base:        "YSO_WIN.exe+2F8314",
				Offsets:     []uint32{},
				Type:        "byte",
				Description: "Current character ID",
			},
			"hp": {
				Base:        "YSO_WIN.exe+2F6ED0",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Current HP",
			},
			"mp": {
				Base:        "YSO_WIN.exe+2F6ED8",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Current MP",
			},
			"experience": {
				Base:        "YSO_WIN.exe+2F6EE8",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Experience points",
			},
			"level": {
				Base:        "YSO_WIN.exe+2F6F00",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "Character level",
			},
			"ground_pointer": {
				Base:        "0x6BB0E0",
				Offsets:     []uint32{0xA10},
				Type:        "byte",
				Description: "Ground state",
			},
			"str_boost": {
				Base:        "YSO_WIN.exe+2F80A4",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "STR boost value",
			},
			"str_boost_visual": {
				Base:        "YSO_WIN.exe+2F80AC",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "STR boost visual timer",
			},
			"def_boost": {
				Base:        "YSO_WIN.exe+2F80A8",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "DEF boost value",
			},
			"def_boost_visual": {
				Base:        "YSO_WIN.exe+2F80B0",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "DEF boost visual timer",
			},
			"mp_boost": {
				Base:        "YSO_WIN.exe+2F80A0",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "MP boost multiplier",
			},
			"mp_boost_visual": {
				Base:        "YSO_WIN.exe+2F80B4",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "MP boost visual timer",
			},
			"hit_count": {
				Base:        "YSO_WIN.exe+2F8948",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "Hit count for EXP boost",
			},
			"hit_count_visual": {
				Base:        "YSO_WIN.exe+2F8950",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "Hit count visual timer",
			},
			"boost_gauge": {
				Base:        "YSO_WIN.exe+2F8954",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Boost gauge value",
			},
			"current_equip_skill": {
				Base:        "YSO_WIN.exe+2F6DD0",
				Offsets:     []uint32{},
				Type:        "byte",
				Description: "Currently equipped skill",
			},
			"floor_boss_hp_gauge": {
				Base:        "0x62CC94",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Floor boss HP gauge",
			},
			"boss_hp_some": {
				Base:        "YSO_WIN.exe+0x25B3D0",
				Offsets:     []uint32{0x4, 0x880},
				Type:        "float32",
				Description: "Some Boss HP",
			},
			"velg_pict_hp": {
				Base:        "YSO_WIN.exe+0x25B460",
				Offsets:     []uint32{0x20, 0x98},
				Type:        "float32",
				Description: "Velg/Pict boss HP",
			},
		},
	},
	"jp_1.0.0.4": {
		Metadata: struct {
			Name        string `yaml:"name"`
			DisplayName string `yaml:"display_name"`
			Description string `yaml:"description"`
			ProcessName string `yaml:"process_name"`
			Game        string `yaml:"game"`
		}{
			Name:        "jp_1.0.0.4",
			DisplayName: "Japanese v1.0.0.4",
			Description: "Ys Origin Japanese version 1.0.0.4 (日本語版)",
			ProcessName: "YSO_WIN.exe",
			Game:        "yso",
		},
		Addresses: map[string]AddressConfig{
			"character": {
				Base:        "YSO_WIN.exe+2E241C",
				Offsets:     []uint32{},
				Type:        "byte",
				Description: "Current character ID",
			},
			"hp": {
				Base:        "YSO_WIN.exe+2E0FD8",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Current HP",
			},
			"mp": {
				Base:        "YSO_WIN.exe+2E0FE0",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Current MP",
			},
			"experience": {
				Base:        "YSO_WIN.exe+2E0FF0",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Experience points",
			},
			"level": {
				Base:        "YSO_WIN.exe+2E1008",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "Character level",
			},
			"ground_pointer": {
				Base:        "YSO_WIN.exe+2A5208",
				Offsets:     []uint32{0xA10},
				Type:        "byte",
				Description: "Ground state",
			},
			"str_boost": {
				Base:        "YSO_WIN.exe+2E21AC",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "STR boost value",
			},
			"str_boost_visual": {
				Base:        "YSO_WIN.exe+2E2A44",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "STR boost visual timer",
			},
			"def_boost": {
				Base:        "YSO_WIN.exe+2E21B0",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "DEF boost value",
			},
			"def_boost_visual": {
				Base:        "YSO_WIN.exe+2E2A48",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "DEF boost visual timer",
			},
			"mp_boost": {
				Base:        "YSO_WIN.exe+2E21A8",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "MP boost value",
			},
			"mp_boost_visual": {
				Base:        "YSO_WIN.exe+2E2A4C",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "MP boost visual timer",
			},
			"hit_count": {
				Base:        "YSO_WIN.exe+2E2A50",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "Hit count for EXP boost",
			},
			"hit_count_visual": {
				Base:        "YSO_WIN.exe+2E2A58",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "Hit count visual timer",
			},
			"boost_gauge": {
				Base:        "YSO_WIN.exe+2E2A5C",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Boost gauge value",
			},
			"current_equip_skill": {
				Base:        "YSO_WIN.exe+2E0ED8",
				Offsets:     []uint32{},
				Type:        "byte",
				Description: "Currently equipped skill",
			},
			"floor_boss_hp_gauge": {
				Base:        "YSO_WIN.exe+21BB94",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "Floor boss HP gauge",
			},
			"boss_hp_some": {
				Base:        "YSO_WIN.exe+0024A314",
				Offsets:     []uint32{0x4, 0x880},
				Type:        "float32",
				Description: "Some Boss HP",
			},
			"velg_pict_hp": {
				Base:        "YSO_WIN.exe+0024A3A8",
				Offsets:     []uint32{0x20, 0x98},
				Type:        "float32",
				Description: "Velg/Pict boss HP",
			},
		},
	},
	"ys6_jp_1.3.1.0": {
		Metadata: struct {
			Name        string `yaml:"name"`
			DisplayName string `yaml:"display_name"`
			Description string `yaml:"description"`
			ProcessName string `yaml:"process_name"`
			Game        string `yaml:"game"`
		}{
			Name:        "ys6_jp_1.3.1.0",
			DisplayName: "Ys VI Japanese v1.3.1.0",
			Description: "Ys VI: The Ark of Napishtim 日本語版 1.3.1.0",
			ProcessName: "YS6_WIN.exe",
			Game:        "ys6",
		},
		Addresses: map[string]AddressConfig{
			"player_base": {
				Base:        "YS6_WIN.exe+177A44",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "プレイヤー構造体ポインタ (0=未ロード)",
			},
			"room_id": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x118},
				Type:        "int32",
				Description: "現在の部屋ID (-1=未ロード, 0=マップ間中継)",
			},
			"level": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x27C},
				Type:        "int32",
				Description: "レベル",
			},
			"experience": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x264},
				Type:        "float32",
				Description: "経験値",
			},
			"sword_livart_level": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x21C},
				Type:        "int32",
				Description: "リヴァート(風/青)剣レベル",
			},
			"sword_brillante_level": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x22C},
				Type:        "int32",
				Description: "ブリランテ(炎/赤)剣レベル",
			},
			"sword_ericcil_level": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x23C},
				Type:        "int32",
				Description: "エリシール(雷)剣レベル",
			},
			"emel": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x45C},
				Type:        "int32",
				Description: "エメル保有数",
			},
			"armor": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x2E8},
				Type:        "int32",
				Description: "装備防具ID (-1=未装備)",
			},
			"shield": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x2D0},
				Type:        "int32",
				Description: "装備盾ID (-1=未装備)",
			},
			"game_mode": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x4D8},
				Type:        "int32",
				Description: "0=ストーリー, 1=タイムアタック",
			},
			"hp": {
				Base:        "YS6_WIN.exe+27C2C",
				Offsets:     []uint32{0x30},
				Type:        "float32",
				Description: "現在HP",
			},
			"hp_visual": {
				Base:        "YS6_WIN.exe+27C2C",
				Offsets:     []uint32{0x34},
				Type:        "float32",
				Description: "表示上のHP",
			},
			"mp_gauge": {
				Base:        "YS6_WIN.exe+177D10",
				Offsets:     []uint32{},
				Type:        "int32",
				Description: "MPゲージ(現在装備の剣)",
			},
			"boss_hp": {
				Base:        "YS6_WIN.exe+13A648",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "ボスHP",
			},
			"boss_max_hp": {
				Base:        "YS6_WIN.exe+13A650",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "ボス最大HP",
			},
			"position_x": {
				Base:        "YS6_WIN.exe+14B284",
				Offsets:     []uint32{0x0, 0x58},
				Type:        "float32",
				Description: "プレイヤーX座標",
			},
			"position_y": {
				Base:        "YS6_WIN.exe+14B284",
				Offsets:     []uint32{0x0, 0x50},
				Type:        "float32",
				Description: "プレイヤーY座標",
			},
			"position_z": {
				Base:        "YS6_WIN.exe+14B284",
				Offsets:     []uint32{0x0, 0x54},
				Type:        "float32",
				Description: "プレイヤーZ座標",
			},
			"equipped_sword": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x430},
				Type:        "int32",
				Description: "装備中の剣 (0=リヴァート(青), 1=ブリランテ(赤), 2=エリシール(雷))",
			},
			"sword_livart_gauge": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x8B0},
				Type:        "float32",
				Description: "リヴァート剣魔法ゲージ (0〜3000、装備中の剣のみ自動回復)",
			},
			"sword_brillante_gauge": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x8B4},
				Type:        "float32",
				Description: "ブリランテ剣魔法ゲージ (0〜3000、装備中の剣のみ自動回復)",
			},
			"sword_ericcil_gauge": {
				Base:        "YS6_WIN.exe+177A44",
				BaseAdd:     "YS6_WIN.exe+177A48",
				Offsets:     []uint32{0x8B8},
				Type:        "float32",
				Description: "エリシール剣魔法ゲージ (0〜3000、装備中の剣のみ自動回復)",
			},
			"max_hp": {
				Base:        "YS6_WIN.exe+13A5C8",
				Offsets:     []uint32{},
				Type:        "float32",
				Description: "プレイヤー最大HP",
			},
			"on_ground": {
				Base:        "YS6_WIN.exe+14B284",
				Offsets:     []uint32{0x0, 0x46C},
				Type:        "int32",
				Description: "接地状態 (1=接地, 0=空中)",
			},
		},
	},
}

// GetAvailableVersions returns all available built-in versions
// mapのrange順は実行のたびに変わるため、対話選択の番号を安定させるためソートして返します。
func GetAvailableVersions() []string {
	versions := make([]string, 0, len(BuiltinVersionConfigs))
	for version := range BuiltinVersionConfigs {
		versions = append(versions, version)
	}
	sort.Strings(versions)
	return versions
}

// GetVersionConfig returns the configuration for a specific version
func GetVersionConfig(version string) (*VersionConfig, bool) {
	config, ok := BuiltinVersionConfigs[version]
	return config, ok
}

// ValidateVersion checks if a version exists
func ValidateVersion(version string) bool {
	_, ok := BuiltinVersionConfigs[version]
	return ok
}
