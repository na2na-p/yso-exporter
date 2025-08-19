package collector

// MetricType は、メトリクスの種類を表す列挙型です。
type MetricType int

const (
	// Gauge は、現在の値を表すメトリクス（例：現在のHP、レベル）
	Gauge MetricType = iota
	// Counter は、累積値を表すメトリクス（例：倒した敵の数、獲得した経験値）
	Counter
)

// String は、MetricTypeの文字列表現を返します。
func (mt MetricType) String() string {
	switch mt {
	case Gauge:
		return "gauge"
	case Counter:
		return "counter"
	default:
		return "unknown"
	}
}

// Metric は、収集されたメトリクスの情報を保持する構造体です。
// カプセル化の原則に従い、メトリクス情報の保持という単一の目的を持ちます。
type Metric struct {
	// Name は、メトリクスの名前です（例：player_hp, player_level）
	Name string
	
	// Value は、メトリクスの値です（数値として表現）
	Value float64
	
	// Type は、メトリクスの種類です（GaugeまたはCounter）
	Type MetricType
	
	// Labels は、メトリクスに付与するラベルのキー・バリューペアです
	// 例：{"character": "yunica", "difficulty": "nightmare"}
	Labels map[string]string
}