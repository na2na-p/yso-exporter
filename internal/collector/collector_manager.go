package collector

import (
	"fmt"
	"log"

	"github.com/na2na/yso-exporter/internal/memory"
	"github.com/na2na/yso-exporter/internal/version"
)

// コンパイル時にCollectorManagerがCollectorインターフェースを実装していることを確認
var _ Collector = (*CollectorManager)(nil)

// CollectorManager は、複数のCollectorを統合管理するCollectorの実装です。
// カプセル化の原則に従い、拡張メトリクス収集の統合という単一の目的を持ちます。
// 関心の分離の原則に従い、各専門Collectorに処理を委譲します。
type CollectorManager struct {
	collectors []Collector
	metrics    []Metric
}

// NewCollectorManager は、CollectorManagerの新しいインスタンスを作成します。
// 引数として渡されたresolver、readerを使用して各専門Collectorを初期化します。
func NewCollectorManager(resolver version.AddressResolver, reader memory.MemoryReader) *CollectorManager {
	collectors := []Collector{
		NewBasicCollector(resolver, reader),
		NewPositionCollector(resolver, reader),
		NewCombatCollector(resolver, reader),
		NewBossCollector(resolver, reader),
		NewFlagCollector(resolver, reader),
	}

	return &CollectorManager{
		collectors: collectors,
		metrics:    make([]Metric, 0),
	}
}

// NewExtendedCollectorManager は、基本コレクターを除く拡張コレクターのみを管理するマネージャーを作成します。
// 基本メトリクスは別のコレクターで管理されている場合に使用します。
func NewExtendedCollectorManager(resolver version.AddressResolver, reader memory.MemoryReader) *CollectorManager {
	collectors := []Collector{
		NewPositionCollector(resolver, reader),
		NewCombatCollector(resolver, reader),
		NewBossCollector(resolver, reader),
		NewFlagCollector(resolver, reader),
	}

	return &CollectorManager{
		collectors: collectors,
		metrics:    make([]Metric, 0),
	}
}

// Collect は、管理している全てのCollectorからメトリクスを収集します。
// 個々のCollectorでエラーが発生しても他のCollectorの実行を継続し、
// 最終的に収集されたメトリクスをすべて統合します。
func (cm *CollectorManager) Collect() error {
	// メトリクスをクリア
	cm.metrics = make([]Metric, 0)
	
	var errors []error

	// 各Collectorからメトリクスを収集
	for i, collector := range cm.collectors {
		err := collector.Collect()
		if err != nil {
			// エラーをログに記録するが、処理は継続
			log.Printf("Collector %d failed to collect metrics: %v", i, err)
			errors = append(errors, fmt.Errorf("collector %d: %w", i, err))
			continue
		}

		// 正常に収集できたメトリクスを統合
		collectedMetrics := collector.GetMetrics()
		cm.metrics = append(cm.metrics, collectedMetrics...)
	}

	// すべてのCollectorが失敗した場合のみエラーを返す
	if len(errors) == len(cm.collectors) {
		return fmt.Errorf("all collectors failed: %v", errors)
	}

	// 一部のCollectorが失敗した場合は警告ログを出力するが成功とする
	if len(errors) > 0 {
		log.Printf("Some collectors failed (%d out of %d): %v", len(errors), len(cm.collectors), errors)
	}

	return nil
}

// GetMetrics は、統合された全てのメトリクスを返します。
func (cm *CollectorManager) GetMetrics() []Metric {
	return cm.metrics
}

// GetCollectorCount は、管理しているCollectorの数を返します。
// テストやデバッグで使用します。
func (cm *CollectorManager) GetCollectorCount() int {
	return len(cm.collectors)
}

// GetCollectorMetrics は、指定されたインデックスのCollectorのメトリクスを個別に取得します。
// デバッグやテスト用途で使用します。
func (cm *CollectorManager) GetCollectorMetrics(index int) ([]Metric, error) {
	if index < 0 || index >= len(cm.collectors) {
		return nil, fmt.Errorf("collector index %d out of range [0, %d)", index, len(cm.collectors))
	}

	return cm.collectors[index].GetMetrics(), nil
}

// CollectIndividually は、各Collectorから個別にメトリクスを収集し、
// Collectorごとの結果をマップで返します。デバッグ用途で使用します。
func (cm *CollectorManager) CollectIndividually() map[string]CollectorResult {
	results := make(map[string]CollectorResult)
	
	collectorNames := []string{"basic", "position", "combat", "boss", "flag"}
	
	for i, collector := range cm.collectors {
		var name string
		if i < len(collectorNames) {
			name = collectorNames[i]
		} else {
			name = fmt.Sprintf("collector_%d", i)
		}

		err := collector.Collect()
		var metrics []Metric
		if err == nil {
			metrics = collector.GetMetrics()
		}

		results[name] = CollectorResult{
			Metrics: metrics,
			Error:   err,
		}
	}

	return results
}

// CollectorResult は、個別のCollectorの収集結果を表します。
type CollectorResult struct {
	Metrics []Metric
	Error   error
}