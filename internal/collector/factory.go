package collector

import (
	"github.com/na2na/yso-exporter/internal/memory"
	"github.com/na2na/yso-exporter/internal/version"
)

// gameYs6 は、Ys VI: The Ark of Napishtim を表すゲーム識別子です。
const gameYs6 = "ys6"

// NewCollectorForGame は、ゲーム識別子に応じたCollectorを生成します。
// game が "ys6" の場合はYs6Collector、それ以外(空文字を含む既存バージョンとの後方互換)の
// 場合はBasicCollectorを返します。
func NewCollectorForGame(game string, resolver version.AddressResolver, reader memory.MemoryReader) Collector {
	if game == gameYs6 {
		return NewYs6Collector(resolver, reader)
	}
	return NewBasicCollector(resolver, reader)
}
