package config

// BuiltinVersionConfigs contains all built-in version configurations
// These are the only supported configurations (no external YAML files)
var BuiltinVersionConfigs = map[string]*VersionConfig{
	"jp_1.1.0.3": {
		Metadata: struct {
			Name        string `yaml:"name"`
			DisplayName string `yaml:"display_name"`
			Description string `yaml:"description"`
			ProcessName string `yaml:"process_name"`
		}{
			Name:        "jp_1.1.0.3",
			DisplayName: "Japanese v1.1.0.3",
			Description: "Ys Origin Japanese version 1.1.0.3 (日本語版)",
			ProcessName: "YSO_WIN.exe",
		},
		Addresses: map[string]AddressConfig{
			"character": {
				Base:        "YSO_WIN.exe+2F8314",
				Offsets:     []uint32{},
				Type:        "byte",
				Description: "Current character ID",
			},
			"hp": {
				Base:        "0x6F6ED0",
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
		}{
			Name:        "jp_1.0.0.4",
			DisplayName: "Japanese v1.0.0.4",
			Description: "Ys Origin Japanese version 1.0.0.4 (日本語版)",
			ProcessName: "YSO_WIN.exe",
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
}

// GetAvailableVersions returns all available built-in versions
func GetAvailableVersions() []string {
	versions := make([]string, 0, len(BuiltinVersionConfigs))
	for version := range BuiltinVersionConfigs {
		versions = append(versions, version)
	}
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
