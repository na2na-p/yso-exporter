package config

import "time"

// AddressConfig represents a single memory address configuration
//
// BaseAdd は任意フィールドです。指定時は Base と BaseAdd がそれぞれ指すアドレスから
// 32bit値を読み取り、両者を加算した値をベースポインタとして扱います。
// (例: LiveSplit ASL の `P = [A] + [B]` のような、2箇所の値を合算して求めるポインタ表現)
type AddressConfig struct {
	Base        interface{} `yaml:"base"`        // Base memory address (can be uint32 or "process.exe+offset" string)
	BaseAdd     interface{} `yaml:"base_add"`    // Optional: address to add to Base (see type doc)
	Offsets     []uint32    `yaml:"offsets"`     // Pointer chain offsets
	Type        string      `yaml:"type"`        // Data type (e.g., "int32", "float32", "string", "byte")
	Description string      `yaml:"description"` // Human-readable description
}

// VersionConfig represents configuration for a specific game version
type VersionConfig struct {
	// Metadata about the game version
	Metadata struct {
		Name        string `yaml:"name"`        // Version name (e.g., "jp_1.1.0.3")
		DisplayName string `yaml:"display_name"` // Human-readable name
		Description string `yaml:"description"` // Version description
		ProcessName string `yaml:"process_name"` // Process executable name
		Game        string `yaml:"game"`         // Game identifier (e.g., "yso", "ys6")
	} `yaml:"metadata"`
	
	// Memory addresses for this version
	Addresses map[string]AddressConfig `yaml:"addresses"`
}

// AppConfig represents the application configuration
type AppConfig struct {
	Version  string        `yaml:"version"`  // Game version to monitor
	Port     int           `yaml:"port"`     // HTTP server port
	Interval time.Duration `yaml:"interval"` // Polling interval
}