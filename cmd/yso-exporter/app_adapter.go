package main

import (
	"fmt"
	"time"

	"github.com/na2na/yso-exporter/internal/config"
	"github.com/na2na/yso-exporter/internal/gui"
	"github.com/na2na/yso-exporter/internal/version"
)

// applicationAdapterImpl implements gui.ApplicationAdapter
// カプセル化: メインアプリケーション適応の実装
type applicationAdapterImpl struct {
	app         *Application
	processName string
	running     bool
}

// Start starts the application
// 単一責任: アプリケーション起動の責任
func (a *applicationAdapterImpl) Start() error {
	if a.app == nil {
		return fmt.Errorf("application not initialized")
	}
	if err := a.app.Start(); err != nil {
		return err
	}
	a.running = true
	return nil
}

// Stop stops the application
// 単一責任: アプリケーション停止の責任
func (a *applicationAdapterImpl) Stop() error {
	if a.app == nil {
		return fmt.Errorf("application not initialized")
	}
	a.app.shutdown()
	a.app.Cleanup()
	a.running = false
	return nil
}

// IsRunning checks if application is running
// カプセル化: 実行状態確認の責任
func (a *applicationAdapterImpl) IsRunning() bool {
	return a.running
}

// GetProcessName returns the target process name
// カプセル化: プロセス名取得の責任
func (a *applicationAdapterImpl) GetProcessName() string {
	return a.processName
}

// applicationLifecycleImpl implements gui.ApplicationLifecycle
// 関心の分離: アプリケーションライフサイクル管理の実装
type applicationLifecycleImpl struct {
	callback ProcessStatusCallback
}

// NewApplicationLifecycle creates a new application lifecycle manager
// 単一責任: ライフサイクル管理の生成
func NewApplicationLifecycle(callback ProcessStatusCallback) gui.ApplicationLifecycle {
	return &applicationLifecycleImpl{
		callback: callback,
	}
}

// Create creates a new application instance
// カプセル化: アプリケーション生成の責任
func (l *applicationLifecycleImpl) Create(versionStr string, port int, interval int) (gui.ApplicationAdapter, error) {
	flags := &config.Flags{
		Version:  versionStr,
		Port:     port,
		Interval: time.Duration(interval) * time.Millisecond,
	}

	// Get process name for the version
	resolver, err := version.NewConfigAddressResolverByVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get resolver: %w", err)
	}
	processName := resolver.GetProcessName()

	// Create application with callback
	app, err := NewApplicationWithCallback(flags, l.callback)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return &applicationAdapterImpl{
		app:         app,
		processName: processName,
		running:     false,
	}, nil
}

// Destroy destroys the application instance
// カプセル化: アプリケーション破棄の責任
func (l *applicationLifecycleImpl) Destroy(adapter gui.ApplicationAdapter) error {
	if adapter == nil {
		return nil
	}
	return adapter.Stop()
}