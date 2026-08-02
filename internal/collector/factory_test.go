package collector_test

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/na2na/yso-exporter/internal/collector"
	memorymocks "github.com/na2na/yso-exporter/internal/memory/mocks"
	versionmocks "github.com/na2na/yso-exporter/internal/version/mocks"
)

func TestNewCollectorForGame(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		game string
		want func(collector.Collector) bool
	}{
		{
			name: "正常系: game=\"ys6\"の場合、Ys6Collectorが返される",
			game: "ys6",
			want: func(c collector.Collector) bool {
				_, ok := c.(*collector.Ys6Collector)
				return ok
			},
		},
		{
			name: "正常系: game=\"yso\"の場合、BasicCollectorが返される",
			game: "yso",
			want: func(c collector.Collector) bool {
				_, ok := c.(*collector.BasicCollector)
				return ok
			},
		},
		{
			name: "正常系: game=\"\"(未指定)の場合、BasicCollectorが返される(後方互換)",
			game: "",
			want: func(c collector.Collector) bool {
				_, ok := c.(*collector.BasicCollector)
				return ok
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			resolver := versionmocks.NewMockAddressResolver(ctrl)
			reader := memorymocks.NewMockMemoryReader(ctrl)

			got := collector.NewCollectorForGame(tt.game, resolver, reader)

			if !tt.want(got) {
				t.Errorf("NewCollectorForGame(%q) returned unexpected type: %T", tt.game, got)
			}
		})
	}
}
