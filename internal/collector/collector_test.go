package collector_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/na2na/yso-exporter/internal/collector"
	"github.com/na2na/yso-exporter/internal/collector/mocks"
	"go.uber.org/mock/gomock"
)

func TestCollector_Collect(t *testing.T) {
	type fields struct {
		collector func(ctrl *gomock.Controller) collector.Collector
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "正常系: メトリクス収集が成功する",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().Collect().Return(nil)
					return m
				},
			},
			wantErr: false,
		},
		{
			name: "異常系: メトリクス収集でエラーが発生する",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().Collect().Return(errors.New("collection failed"))
					return m
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			
			// 準備: モックをセットアップ
			c := tt.fields.collector(ctrl)
			
			// 実行: テスト対象メソッドを呼び出し
			err := c.Collect()
			
			// 検証: エラーの有無を比較
			if (err != nil) != tt.wantErr {
				t.Errorf("Collector.Collect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollector_GetMetrics(t *testing.T) {
	type fields struct {
		collector func(ctrl *gomock.Controller) collector.Collector
	}
	tests := []struct {
		name   string
		fields fields
		want   []collector.Metric
	}{
		{
			name: "正常系: 空のメトリクスリストが返される",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().GetMetrics().Return([]collector.Metric{})
					return m
				},
			},
			want: []collector.Metric{},
		},
		{
			name: "正常系: メトリクスリストが返される",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					expectedMetrics := []collector.Metric{
						{
							Name:   "player_hp",
							Value:  100.0,
							Type:   collector.Gauge,
							Labels: map[string]string{"character": "yunica"},
						},
						{
							Name:   "enemies_defeated",
							Value:  5.0,
							Type:   collector.Counter,
							Labels: map[string]string{"area": "tower"},
						},
					}
					m.EXPECT().GetMetrics().Return(expectedMetrics)
					return m
				},
			},
			want: []collector.Metric{
				{
					Name:   "player_hp",
					Value:  100.0,
					Type:   collector.Gauge,
					Labels: map[string]string{"character": "yunica"},
				},
				{
					Name:   "enemies_defeated",
					Value:  5.0,
					Type:   collector.Counter,
					Labels: map[string]string{"area": "tower"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			
			// 準備: モックをセットアップ
			c := tt.fields.collector(ctrl)
			
			// 実行: テスト対象メソッドを呼び出し
			got := c.GetMetrics()
			
			// 検証: 戻り値を比較
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Collector.GetMetrics() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}