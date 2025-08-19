package collector

import (
	"context"
	"fmt"
	"sync"
	"time"
)

//go:generate mockgen -source=poller.go -destination=mocks/mock_poller.go -package=mocks

// MetricPoller は、定期的にメトリクスを収集するためのインターフェースです。
// カプセル化の原則に従い、ポーリングによるメトリクス収集という単一の目的を持ちます。
type MetricPoller interface {
	// Start は、指定した間隔でポーリングを開始します
	// contextがキャンセルされるまでポーリングを継続します
	Start(ctx context.Context, interval time.Duration) error
	
	// Stop は、ポーリングを停止します
	// Start()で開始されたゴルーチンが確実に終了するまで待機します
	Stop() error
	
	// IsRunning は、ポーリングが実行中かどうかを返します
	IsRunning() bool
}

// PollerImpl は、MetricPollerインターフェースの実装です。
// カプセル化の原則に従い、ポーリング処理の管理という単一の目的を持ちます。
type PollerImpl struct {
	collector Collector
	mu        sync.RWMutex
	isRunning bool
	cancel    context.CancelFunc
	done      chan struct{}
}

// NewPoller は、PollerImplのコンストラクタです。
func NewPoller(collector Collector) *PollerImpl {
	return &PollerImpl{
		collector: collector,
		done:      make(chan struct{}),
	}
}

// Start は、指定した間隔でポーリングを開始します。
func (p *PollerImpl) Start(ctx context.Context, interval time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.isRunning {
		return fmt.Errorf("poller is already running")
	}
	
	if interval <= 0 {
		return fmt.Errorf("interval must be positive: %v", interval)
	}
	
	pollerCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.isRunning = true
	
	go p.pollLoop(pollerCtx, interval)
	
	return nil
}

// Stop は、ポーリングを停止します。
func (p *PollerImpl) Stop() error {
	p.mu.Lock()
	
	if !p.isRunning {
		p.mu.Unlock()
		return fmt.Errorf("poller is not running")
	}
	
	// キャンセルを実行
	if p.cancel != nil {
		p.cancel()
	}
	
	p.mu.Unlock()
	
	// ポーリングループの終了を待機（タイムアウト付き）
	select {
	case <-p.done:
		// 正常に終了
	case <-time.After(1 * time.Second):
		// タイムアウト（通常は発生しないはず）
	}
	
	p.mu.Lock()
	p.isRunning = false
	p.cancel = nil
	p.mu.Unlock()
	
	return nil
}

// IsRunning は、ポーリングが実行中かどうかを返します。
func (p *PollerImpl) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isRunning
}

// pollLoop は、実際のポーリング処理を行うプライベートメソッドです。
// 関心の分離の原則に従い、ポーリングのコア処理という単一の目的を持ちます。
func (p *PollerImpl) pollLoop(ctx context.Context, interval time.Duration) {
	defer func() {
		// 非ブロッキングでdoneチャネルに送信
		select {
		case p.done <- struct{}{}:
		default:
			// チャネルがフルの場合はスキップ
		}
	}()
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// メトリクス収集を実行
			// エラーが発生してもポーリングは継続する
			_ = p.collector.Collect()
		}
	}
}