package exporter

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/na2na/yso-exporter/internal/collector"
)

//go:generate mockgen -source=prometheus.go -destination=mocks/mock_exporter.go -package=mocks

// Exporter は、メトリクスをエクスポートするためのインターフェースです。
// カプセル化の原則に従い、メトリクスエクスポートという単一の目的を持ちます。
type Exporter interface {
	// ServeHTTP は、HTTPリクエストを処理してメトリクスを返します
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// PrometheusExporter は、Prometheus形式でメトリクスをエクスポートする実装です。
// 関心の分離の原則に従い、Prometheus形式での出力に特化します。
type PrometheusExporter struct {
	collector collector.Collector
}

// NewPrometheusExporter は、PrometheusExporterの新しいインスタンスを作成します。
func NewPrometheusExporter(collector collector.Collector) *PrometheusExporter {
	return &PrometheusExporter{
		collector: collector,
	}
}

// ServeHTTP は、HTTPリクエストを処理してPrometheus形式のメトリクスを返します。
func (pe *PrometheusExporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// HTTPメソッドの検証
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed")
		return
	}

	// Content-Typeヘッダーを設定
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	// メトリクス収集はPollerが単独で行う。ハンドラは収集済みスナップショットを
	// 読むだけにすることで、収集と参照の同時実行による競合を避ける。
	metrics := pe.collector.GetMetrics()

	// Prometheus形式で出力
	pe.writePrometheusFormat(w, metrics)
}

// writePrometheusFormat は、メトリクスをPrometheus形式でWriterに出力します。
func (pe *PrometheusExporter) writePrometheusFormat(w http.ResponseWriter, metrics []collector.Metric) {
	// メトリクス名でグループ化
	metricGroups := make(map[string][]collector.Metric)
	for _, metric := range metrics {
		metricGroups[metric.Name] = append(metricGroups[metric.Name], metric)
	}

	// メトリクス名をソートして一貫した出力順序を保証
	names := make([]string, 0, len(metricGroups))
	for name := range metricGroups {
		names = append(names, name)
	}
	sort.Strings(names)

	// 各メトリクスグループを出力
	for _, name := range names {
		group := metricGroups[name]
		if len(group) == 0 {
			continue
		}

		// HELP行を出力（メトリクス名の説明）
		fmt.Fprintf(w, "# HELP %s %s\n", name, name)

		// TYPE行を出力（メトリクスの種類）
		fmt.Fprintf(w, "# TYPE %s %s\n", name, group[0].Type.String())

		// 各メトリクスの値を出力
		for _, metric := range group {
			pe.writeMetricLine(w, metric)
		}
	}
}

// writeMetricLine は、単一のメトリクス行をPrometheus形式で出力します。
func (pe *PrometheusExporter) writeMetricLine(w http.ResponseWriter, metric collector.Metric) {
	if len(metric.Labels) == 0 {
		// ラベルなしの場合
		fmt.Fprintf(w, "%s %g\n", metric.Name, metric.Value)
		return
	}

	// ラベルありの場合
	labels := pe.formatLabels(metric.Labels)
	fmt.Fprintf(w, "%s%s %g\n", metric.Name, labels, metric.Value)
}

// formatLabels は、ラベルマップをPrometheus形式の文字列に変換します。
func (pe *PrometheusExporter) formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	// キーをソートして一貫した出力順序を保証
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// ラベル文字列を構築
	var parts []string
	for _, key := range keys {
		value := labels[key]
		// ラベル値をエスケープ（簡易版）
		escapedValue := strings.ReplaceAll(value, "\"", "\\\"")
		parts = append(parts, fmt.Sprintf("%s=\"%s\"", key, escapedValue))
	}

	return "{" + strings.Join(parts, ",") + "}"
}
