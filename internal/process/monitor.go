package process

import (
	"context"
	"errors"
	"sync"
	"time"
)

//go:generate mockgen -source=monitor.go -destination=mocks/mock_monitor.go -package=mocks

// ProcessMonitor は、指定されたプロセスの監視を行うインターフェースです。
// カプセル化の原則に従い、プロセス監視という単一の目的を持ちます。
type ProcessMonitor interface {
	// Start は指定されたプロセス名の監視を開始します
	Start(ctx context.Context, processName string) error
	
	// Stop は監視を停止します
	Stop() error
	
	// IsProcessAlive は現在プロセスが生きているかを確認します
	IsProcessAlive() bool
}

// ProcessEventHandler は、プロセスの状態変化を処理するためのインターフェースです。
// 関心の分離の原則に従い、プロセス監視とイベント処理を分離します。
type ProcessEventHandler interface {
	// OnProcessFound はプロセスが見つかったときに呼び出されます
	OnProcessFound(processName string) error
	
	// OnProcessLost はプロセスが失われたときに呼び出されます
	OnProcessLost(processName string) error
}

// MonitorConfig は、プロセス監視の設定を定義します。
type MonitorConfig struct {
	// MonitorInterval は監視間隔を指定します（デフォルト: 1秒）
	MonitorInterval time.Duration
	
	// EventHandler はプロセス状態変化を処理するハンドラです
	EventHandler ProcessEventHandler
}

// ProcessMonitorImpl は、ProcessMonitorインターフェースの実装です。
// カプセル化の原則に従い、プロセス監視の責任のみを持ちます。
type ProcessMonitorImpl struct {
	config        *MonitorConfig
	processName   string
	isAlive       bool
	isRunning     bool
	stopCh        chan struct{}
	mu            sync.RWMutex
	processFinder ProcessFinder
}

// NewProcessMonitor は、新しいProcessMonitorを作成します。
func NewProcessMonitor(config *MonitorConfig, finder ProcessFinder) ProcessMonitor {
	if config.MonitorInterval == 0 {
		config.MonitorInterval = 1 * time.Second
	}
	
	return &ProcessMonitorImpl{
		config:        config,
		stopCh:        make(chan struct{}),
		processFinder: finder,
	}
}

// Start は指定されたプロセス名の監視を開始します。
func (m *ProcessMonitorImpl) Start(ctx context.Context, processName string) error {
	if processName == "" {
		return errors.New("process name cannot be empty")
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.isRunning {
		return errors.New("monitor is already running")
	}
	
	m.processName = processName
	m.isRunning = true
	
	go m.monitorLoop(ctx)
	
	return nil
}

// Stop は監視を停止します。
func (m *ProcessMonitorImpl) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.isRunning {
		return nil
	}
	
	close(m.stopCh)
	m.isRunning = false
	m.isAlive = false
	
	// 新しいstopチャンネルを作成
	m.stopCh = make(chan struct{})
	
	return nil
}

// IsProcessAlive は現在プロセスが生きているかを確認します。
func (m *ProcessMonitorImpl) IsProcessAlive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isAlive
}

// monitorLoop は、プロセス監視のメインループです。
func (m *ProcessMonitorImpl) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(m.config.MonitorInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkProcess()
		}
	}
}

// checkProcess は、プロセスの存在を確認し、状態変化を処理します。
func (m *ProcessMonitorImpl) checkProcess() {
	processExists := m.isProcessExists(m.processName)
	
	m.mu.Lock()
	previousState := m.isAlive
	m.isAlive = processExists
	m.mu.Unlock()
	
	// 状態変化の検知と処理
	if !previousState && processExists {
		// プロセスが見つかった
		if m.config.EventHandler != nil {
			if err := m.config.EventHandler.OnProcessFound(m.processName); err != nil {
				// エラーログなどの処理は呼び出し側に委ねる
			}
		}
	} else if previousState && !processExists {
		// プロセスが失われた
		if m.config.EventHandler != nil {
			if err := m.config.EventHandler.OnProcessLost(m.processName); err != nil {
				// エラーログなどの処理は呼び出し側に委ねる
			}
		}
	}
}

// isProcessExists は、指定されたプロセス名のプロセスが存在するかを確認します。
func (m *ProcessMonitorImpl) isProcessExists(processName string) bool {
	_, err := m.processFinder.FindProcessByName(processName)
	return err == nil
}