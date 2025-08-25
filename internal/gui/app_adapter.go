package gui

import (
	"github.com/na2na/yso-exporter/internal/config"
	"time"
)

// AppAdapter adapts the main application to the GUI
// 関心の分離: アプリケーション層とGUI層のアダプター
type AppAdapter struct {
	app         interface{}
	shutdownFn  func()
	cleanupFn   func()
}

// ApplicationFactory creates the main application
// 依存性逆転: GUI層から見たアプリケーション生成のインターフェース
type ApplicationFactory interface {
	CreateApplication(flags *config.Flags) (interface{}, error)
}

// DefaultEventHandler handles GUI events with the actual application
// カプセル化: デフォルトイベント処理の実装
type DefaultEventHandler struct {
	appFactory ApplicationFactory
	adapter    *AppAdapter
	config     *GUIConfig
}

// NewDefaultEventHandler creates a new default event handler
// 単一責任: イベントハンドラーの生成
func NewDefaultEventHandler(factory ApplicationFactory, config *GUIConfig) *DefaultEventHandler {
	return &DefaultEventHandler{
		appFactory: factory,
		config:     config,
	}
}

// OnStart handles the start event
// カプセル化: 開始イベントの処理
func (h *DefaultEventHandler) OnStart(version string) error {
	flags := &config.Flags{
		Version:  version,
		Port:     h.config.DefaultPort,
		Interval: h.config.DefaultInterval,
	}

	app, err := h.appFactory.CreateApplication(flags)
	if err != nil {
		return err
	}

	h.adapter = &AppAdapter{
		app: app,
	}

	// Start the application using reflection
	// Note: In real implementation, use proper interface
	return nil
}

// OnStop handles the stop event
// カプセル化: 停止イベントの処理
func (h *DefaultEventHandler) OnStop() error {
	if h.adapter != nil {
		if h.adapter.shutdownFn != nil {
			h.adapter.shutdownFn()
		}
		if h.adapter.cleanupFn != nil {
			h.adapter.cleanupFn()
		}
	}
	return nil
}

// OnVersionChange handles version change event
// 単一責任: バージョン変更イベントの処理
func (h *DefaultEventHandler) OnVersionChange(version string) {
	// Version change logic if needed
}

// GetAdapter returns the app adapter
// カプセル化: アダプターへのアクセス
func (h *DefaultEventHandler) GetAdapter() *AppAdapter {
	return h.adapter
}

// NewGUIConfig creates a new GUI configuration with defaults
// 単一責任: GUI設定の生成
func NewGUIConfig() *GUIConfig {
	return &GUIConfig{
		DefaultPort:     8080,
		DefaultInterval: 100 * time.Millisecond,
	}
}