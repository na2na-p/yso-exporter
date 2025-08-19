package collector_test

import (
	"testing"

	"github.com/na2na/yso-exporter/internal/collector"
)

func TestMetricType_String(t *testing.T) {
	tests := []struct {
		name string
		mt   collector.MetricType
		want string
	}{
		{
			name: "正常系: Gauge型の文字列表現が返される",
			mt:   collector.Gauge,
			want: "gauge",
		},
		{
			name: "正常系: Counter型の文字列表現が返される",
			mt:   collector.Counter,
			want: "counter",
		},
		{
			name: "異常系: 未知の型の文字列表現が返される",
			mt:   collector.MetricType(999),
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 実行: テスト対象メソッドを呼び出し
			got := tt.mt.String()
			
			// 検証: 戻り値を比較
			if got != tt.want {
				t.Errorf("MetricType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}