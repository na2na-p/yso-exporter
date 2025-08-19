package collector

//go:generate mockgen -source=collector.go -destination=mocks/mock_collector.go -package=mocks

// Collector は、ゲームメトリクスを収集するためのインターフェースです。
// カプセル化の原則に従い、メトリクス収集という単一の目的を持ちます。
type Collector interface {
	// Collect は、現在のゲーム状態からメトリクスを収集します
	Collect() error
	
	// GetMetrics は、収集済みのメトリクスを取得します
	GetMetrics() []Metric
}