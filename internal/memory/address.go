package memory

import (
	"fmt"
	"strconv"
	"strings"
)

// Address はメモリアドレスを表すValue Objectです
// カプセル化の原則に従い、アドレスに関する全ての知識と操作をカプセル化します
type Address struct {
	value uintptr
}

// NewAddress は生のアドレス値からAddressを作成します
func NewAddress(value uintptr) Address {
	return Address{value: value}
}

// NewAddressFromInterface は様々な形式のアドレスからAddressを作成します
// この関数は型変換のみを責任とし、実際の解析は専門のパーサーに委譲します
func NewAddressFromInterface(addr interface{}, resolver ModuleResolver) (Address, error) {
	switch v := addr.(type) {
	case uint32:
		return NewAddress(uintptr(v)), nil
	case uint64:
		return NewAddress(uintptr(v)), nil
	case int:
		return NewAddress(uintptr(v)), nil
	case int64:
		return NewAddress(uintptr(v)), nil
	case float64:
		// YAML may parse hex numbers as float64
		return NewAddress(uintptr(uint32(v))), nil
	case string:
		parser := NewAddressParser(resolver)
		return parser.Parse(v)
	default:
		return Address{}, fmt.Errorf("unsupported address type: %T", addr)
	}
}

// Value はアドレス値を返します
func (a Address) Value() uintptr {
	return a.value
}

// AddressParser は文字列形式のアドレスを解析する責任を持ちます
// 単一の目的: 文字列をAddressに変換すること
type AddressParser struct {
	moduleResolver ModuleResolver
	formatParsers  []FormatParser
}

// NewAddressParser は新しいAddressParserを作成します
func NewAddressParser(resolver ModuleResolver) *AddressParser {
	return &AddressParser{
		moduleResolver: resolver,
		formatParsers: []FormatParser{
			&ModuleOffsetParser{resolver: resolver},
			&CalculationParser{},
			&PlainAddressParser{},
		},
	}
}

// Parse は文字列をAddressに変換します
func (p *AddressParser) Parse(s string) (Address, error) {
	s = strings.TrimSpace(s)
	
	for _, parser := range p.formatParsers {
		if parser.CanParse(s) {
			return parser.Parse(s)
		}
	}
	
	return Address{}, fmt.Errorf("unsupported address format: %s", s)
}

// FormatParser は特定フォーマットのアドレス解析を責任とするインターフェース
type FormatParser interface {
	CanParse(s string) bool
	Parse(s string) (Address, error)
}

// ModuleOffsetParser は "module.exe+offset" 形式のアドレスを解析します
// 単一の目的: モジュール相対アドレスの解析
type ModuleOffsetParser struct {
	resolver ModuleResolver
}

func (p *ModuleOffsetParser) CanParse(s string) bool {
	return strings.Contains(s, "+")
}

func (p *ModuleOffsetParser) Parse(s string) (Address, error) {
	parts := strings.Split(s, "+")
	if len(parts) != 2 {
		return Address{}, fmt.Errorf("invalid module+offset format: %s", s)
	}
	
	moduleName := strings.TrimSpace(parts[0])
	offsetStr := strings.TrimSpace(parts[1])
	
	// モジュールベースアドレスの取得は専門のリゾルバーに委譲
	moduleBase, err := p.resolver.GetModuleBase(moduleName)
	if err != nil {
		return Address{}, fmt.Errorf("failed to get module base for %s: %w", moduleName, err)
	}
	
	// オフセットの解析は専門のパーサーに委譲
	offset, err := ParseHexOrDecimal(offsetStr)
	if err != nil {
		return Address{}, fmt.Errorf("failed to parse offset %s: %w", offsetStr, err)
	}
	
	return NewAddress(moduleBase + uintptr(offset)), nil
}

// CalculationParser は "address-offset" 形式のアドレスを解析します
// 単一の目的: 計算式形式のアドレスの解析
type CalculationParser struct{}

func (p *CalculationParser) CanParse(s string) bool {
	return strings.Contains(s, "-")
}

func (p *CalculationParser) Parse(s string) (Address, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return Address{}, fmt.Errorf("invalid calculation format: %s", s)
	}
	
	base, err := ParseHexOrDecimal(strings.TrimSpace(parts[0]))
	if err != nil {
		return Address{}, fmt.Errorf("failed to parse base %s: %w", parts[0], err)
	}
	
	subtract, err := ParseHexOrDecimal(strings.TrimSpace(parts[1]))
	if err != nil {
		return Address{}, fmt.Errorf("failed to parse subtraction %s: %w", parts[1], err)
	}
	
	return NewAddress(uintptr(base - subtract)), nil
}

// PlainAddressParser は単純な16進数/10進数アドレスを解析します
// 単一の目的: プレーンなアドレス値の解析
type PlainAddressParser struct{}

func (p *PlainAddressParser) CanParse(s string) bool {
	// 最後のフォールバックパーサーなので常にtrue
	return true
}

func (p *PlainAddressParser) Parse(s string) (Address, error) {
	value, err := ParseHexOrDecimal(s)
	if err != nil {
		return Address{}, fmt.Errorf("failed to parse address %s: %w", s, err)
	}
	
	return NewAddress(uintptr(value)), nil
}

// ParseHexOrDecimal は16進数または10進数の文字列を解析します
// 単一の目的: 数値文字列のパース
func ParseHexOrDecimal(s string) (uint32, error) {
	s = strings.TrimSpace(s)
	
	// Remove common hex prefixes
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	s = strings.TrimPrefix(s, "$")
	
	// Try hex first (most addresses are hex)
	if value, err := strconv.ParseUint(s, 16, 32); err == nil {
		return uint32(value), nil
	}
	
	// Fall back to decimal
	if value, err := strconv.ParseUint(s, 10, 32); err == nil {
		return uint32(value), nil
	}
	
	return 0, fmt.Errorf("invalid number format: %s", s)
}