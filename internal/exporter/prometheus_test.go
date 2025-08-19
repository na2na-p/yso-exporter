package exporter_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/na2na/yso-exporter/internal/collector"
	"github.com/na2na/yso-exporter/internal/collector/mocks"
	"github.com/na2na/yso-exporter/internal/exporter"
)

func TestNewPrometheusExporter(t *testing.T) {
	type args struct {
		collector collector.Collector
	}
	tests := []struct {
		name string
		args args
		want *exporter.PrometheusExporter
	}{
		{
			name: "正常系: 有効なCollectorでPrometheusExporterが作成される",
			args: args{
				collector: &mocks.MockCollector{},
			},
			want: &exporter.PrometheusExporter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exporter.NewPrometheusExporter(tt.args.collector)
			if got == nil {
				t.Error("NewPrometheusExporter() returned nil")
			}
		})
	}
}

func TestPrometheusExporter_ServeHTTP(t *testing.T) {
	type fields struct {
		collector func(ctrl *gomock.Controller) collector.Collector
	}
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantStatusCode int
		wantHeaders    map[string]string
		wantBodyContains []string
		wantErr        error
	}{
		{
			name: "正常系: Gaugeメトリクスが正しいPrometheus形式で出力される",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().Collect().Return(nil)
					m.EXPECT().GetMetrics().Return([]collector.Metric{
						{
							Name:   "player_hp",
							Value:  100.0,
							Type:   collector.Gauge,
							Labels: map[string]string{"character": "yunica"},
						},
					})
					return m
				},
			},
			args: args{
				method: http.MethodGet,
				path:   "/metrics",
			},
			wantStatusCode: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type": "text/plain; version=0.0.4; charset=utf-8",
			},
			wantBodyContains: []string{
				"# HELP player_hp",
				"# TYPE player_hp gauge",
				"player_hp{character=\"yunica\"} 100",
			},
		},
		{
			name: "正常系: Counterメトリクスが正しいPrometheus形式で出力される",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().Collect().Return(nil)
					m.EXPECT().GetMetrics().Return([]collector.Metric{
						{
							Name:   "enemies_defeated_total",
							Value:  50.0,
							Type:   collector.Counter,
							Labels: map[string]string{"level": "tower1"},
						},
					})
					return m
				},
			},
			args: args{
				method: http.MethodGet,
				path:   "/metrics",
			},
			wantStatusCode: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type": "text/plain; version=0.0.4; charset=utf-8",
			},
			wantBodyContains: []string{
				"# HELP enemies_defeated_total",
				"# TYPE enemies_defeated_total counter",
				"enemies_defeated_total{level=\"tower1\"} 50",
			},
		},
		{
			name: "正常系: 複数のメトリクスが正しく出力される",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().Collect().Return(nil)
					m.EXPECT().GetMetrics().Return([]collector.Metric{
						{
							Name:   "player_hp",
							Value:  100.0,
							Type:   collector.Gauge,
							Labels: map[string]string{"character": "yunica"},
						},
						{
							Name:   "player_sp",
							Value:  80.0,
							Type:   collector.Gauge,
							Labels: map[string]string{"character": "yunica"},
						},
					})
					return m
				},
			},
			args: args{
				method: http.MethodGet,
				path:   "/metrics",
			},
			wantStatusCode: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type": "text/plain; version=0.0.4; charset=utf-8",
			},
			wantBodyContains: []string{
				"player_hp{character=\"yunica\"} 100",
				"player_sp{character=\"yunica\"} 80",
			},
		},
		{
			name: "正常系: ラベルなしメトリクスが正しく出力される",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().Collect().Return(nil)
					m.EXPECT().GetMetrics().Return([]collector.Metric{
						{
							Name:   "game_version",
							Value:  1.0,
							Type:   collector.Gauge,
							Labels: nil,
						},
					})
					return m
				},
			},
			args: args{
				method: http.MethodGet,
				path:   "/metrics",
			},
			wantStatusCode: http.StatusOK,
			wantHeaders: map[string]string{
				"Content-Type": "text/plain; version=0.0.4; charset=utf-8",
			},
			wantBodyContains: []string{
				"game_version 1",
			},
		},
		{
			name: "異常系: Collector.Collect()でエラーが発生した場合、500エラーが返る",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().Collect().Return(http.ErrServerClosed)
					return m
				},
			},
			args: args{
				method: http.MethodGet,
				path:   "/metrics",
			},
			wantStatusCode: http.StatusInternalServerError,
			wantHeaders: map[string]string{
				"Content-Type": "text/plain; charset=utf-8",
			},
			wantBodyContains: []string{
				"Error collecting metrics",
			},
		},
		{
			name: "異常系: POST方法では405エラーが返る",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					// POST方法では呼び出されない
					return m
				},
			},
			args: args{
				method: http.MethodPost,
				path:   "/metrics",
			},
			wantStatusCode: http.StatusMethodNotAllowed,
			wantHeaders: map[string]string{
				"Content-Type": "text/plain; charset=utf-8",
			},
			wantBodyContains: []string{
				"Method not allowed",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			// 準備: モックをセットアップし、テスト対象インスタンスを生成
			pe := exporter.NewPrometheusExporter(tt.fields.collector(ctrl))

			// HTTPリクエストとレスポンスレコーダーを準備
			req := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			w := httptest.NewRecorder()

			// 実行: テスト対象メソッドを呼び出し
			pe.ServeHTTP(w, req)

			// 検証: ステータスコードを確認
			if diff := cmp.Diff(tt.wantStatusCode, w.Code); diff != "" {
				t.Errorf("ServeHTTP() status code mismatch (-want +got):\n%s", diff)
			}

			// 検証: ヘッダーを確認
			for key, want := range tt.wantHeaders {
				got := w.Header().Get(key)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("ServeHTTP() header %s mismatch (-want +got):\n%s", key, diff)
				}
			}

			// 検証: レスポンスボディの内容を確認
			body := w.Body.String()
			for _, want := range tt.wantBodyContains {
				if !strings.Contains(body, want) {
					t.Errorf("ServeHTTP() body does not contain %q, got: %s", want, body)
				}
			}
		})
	}
}