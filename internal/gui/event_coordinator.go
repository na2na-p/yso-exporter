package gui

import (
	"fmt"
	"log"
)

// EventCoordinator coordinates events between GUI and application
// カプセル化: イベント調整の単一責任
type EventCoordinator struct {
	controller  *Controller
	lifecycle   ApplicationLifecycle
	adapter     ApplicationAdapter
	notifier    *ProcessNotifier
}

// NewEventCoordinator creates a new event coordinator
// 単一責任: コーディネーター生成
func NewEventCoordinator(controller *Controller, lifecycle ApplicationLifecycle) *EventCoordinator {
	return &EventCoordinator{
		controller: controller,
		lifecycle:  lifecycle,
	}
}

// SetNotifier sets the process notifier
// 依存性注入: 通知機能の注入
func (e *EventCoordinator) SetNotifier(notifier *ProcessNotifier) {
	e.notifier = notifier
}

// OnStart handles the start event
// カプセル化: 開始イベント処理の責任
func (e *EventCoordinator) OnStart(version string) error {
	if e.adapter != nil && e.adapter.IsRunning() {
		return fmt.Errorf("already running")
	}

	config := e.controller.config
	adapter, err := e.lifecycle.Create(version, config.DefaultPort, int(config.DefaultInterval.Milliseconds()))
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	e.adapter = adapter

	// Start in goroutine to avoid blocking
	go func() {
		if err := adapter.Start(); err != nil {
			log.Printf("Failed to start application: %v", err)
			e.controller.Stop()
		}
	}()

	return nil
}

// OnStop handles the stop event
// カプセル化: 停止イベント処理の責任
func (e *EventCoordinator) OnStop() error {
	if e.adapter == nil || !e.adapter.IsRunning() {
		return fmt.Errorf("not running")
	}

	if err := e.lifecycle.Destroy(e.adapter); err != nil {
		return fmt.Errorf("failed to destroy application: %w", err)
	}

	e.adapter = nil
	return nil
}

// OnVersionChange handles version change
// 単一責任: バージョン変更処理
func (e *EventCoordinator) OnVersionChange(version string) {
	// No action needed while running
}

// OnProcessConnected handles process connection
// カプセル化: プロセス接続通知の責任
func (e *EventCoordinator) OnProcessConnected(processName string) {
	log.Printf("Process connected: %s", processName)
	if e.controller != nil {
		e.controller.UpdateProcessStatus(true, processName)
	}
}

// OnProcessLost handles process disconnection
// カプセル化: プロセス切断通知の責任
func (e *EventCoordinator) OnProcessLost(processName string) {
	log.Printf("Process lost: %s", processName)
	if e.controller != nil {
		e.controller.UpdateProcessStatus(false, processName)
	}
}