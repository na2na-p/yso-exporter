package collector

import "sync"

// コンパイル時にSnapshotCollectorがCollectorインターフェースを実装していることを確認
var _ Collector = (*SnapshotCollector)(nil)

// SnapshotCollector は、内側のCollectorをラップし、収集結果をRWMutexで保護された
// スナップショットとして保持するデコレータです。収集(書き込み)と参照(読み取り)を
// 別々のゴルーチンから同時に呼んでも内側の可変状態(メトリクススライス)へ
// 直接アクセスさせないという単一の目的を持ちます。
type SnapshotCollector struct {
	inner Collector

	// collectMu は、inner.Collect()〜inner.GetMetrics()の区間を直列化します。
	// Collectの呼び出し元が現在Poller1本に限られているという不変条件は配線側の都合でしか
	// 保証されておらず、呼び出し元が増えた場合に内側の可変状態への競合が再発するため、
	// 型自身がこの不変条件を守ります。スナップショットの読み書きロック(mu)とは別にすることで、
	// GetMetrics()(読み手)がメモリ読み取り中のCollect()の完了を待たされないようにしています。
	collectMu sync.Mutex

	mu       sync.RWMutex
	snapshot []Metric
}

// NewSnapshotCollector は、SnapshotCollectorの新しいインスタンスを作成します。
func NewSnapshotCollector(inner Collector) *SnapshotCollector {
	return &SnapshotCollector{
		inner: inner,
	}
}

// Collect は、内側のCollectorで収集を行い、成功した場合のみスナップショットを更新します。
// 内側がエラーを返した場合はスナップショットを空にしてエラーをそのまま返します。
// 直前の値を保持し続けると、ゲーム終了後などの本物の収集不能時にも古い値を配信し続けてしまうため、
// 収集エラー時は必ず空を返す側に倒します(直前値保持が必要なメトリクスはCollector実装側の責務とする)。
func (sc *SnapshotCollector) Collect() error {
	sc.collectMu.Lock()
	defer sc.collectMu.Unlock()

	if err := sc.inner.Collect(); err != nil {
		sc.mu.Lock()
		sc.snapshot = nil
		sc.mu.Unlock()
		return err
	}

	metrics := copyMetrics(sc.inner.GetMetrics())

	sc.mu.Lock()
	sc.snapshot = metrics
	sc.mu.Unlock()

	return nil
}

// GetMetrics は、直近のスナップショットのコピーを返します。
// 呼び出し側が戻り値を変更しても内部状態に影響しません。
func (sc *SnapshotCollector) GetMetrics() []Metric {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return copyMetrics(sc.snapshot)
}

// copyMetrics は、Metricスライスの独立したコピーを作成します。
// LabelsはGoのマップ(参照型)のため、要素単位で複製しないと呼び出し側の変更が
// スライスコピーを飛び越えて内部状態へ波及してしまいます。
func copyMetrics(src []Metric) []Metric {
	if src == nil {
		return nil
	}

	dst := make([]Metric, len(src))
	for i, m := range src {
		dst[i] = m
		if m.Labels != nil {
			labels := make(map[string]string, len(m.Labels))
			for k, v := range m.Labels {
				labels[k] = v
			}
			dst[i].Labels = labels
		}
	}
	return dst
}
