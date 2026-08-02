package gui

import (
	"fmt"
	"time"
)

// Controller manages the GUI application state and business logic
// カプセル化: GUI制御の単一責任を持つ
type Controller struct {
	app             Application
	config          *GUIConfig
	state           *State
	eventHandler    EventHandler
}

// Application interface for main application
// 関心の分離: アプリケーション層とGUI層の分離
type Application interface {
	Start() error
	Shutdown()
	Cleanup()
}

// EventHandler handles GUI events
// カプセル化: イベント処理の責任を分離
type EventHandler interface {
	OnStart(version string) error
	OnStop() error
	OnVersionChange(version string)
	OnProcessConnected(processName string)
	OnProcessLost(processName string)
	OnExit()
}

// GUIConfig holds GUI-specific configuration
// カプセル化: GUI設定の責任を分離
type GUIConfig struct {
	DefaultPort     int
	DefaultInterval time.Duration
	Versions        []string
}

// State represents the current GUI state
// カプセル化: 状態管理の責任を分離
type State struct {
	Running         bool
	SelectedVersion string
	StatusText      string
	MetricsURL      string
	ProcessStatus   string
	ProcessConnected bool
}

// NewController creates a new GUI controller
// 単一責任: コントローラーの初期化のみ
func NewController(config *GUIConfig) *Controller {
	return &Controller{
		config: config,
		state: &State{
			Running: false,
			ProcessStatus: "Process: Not connected",
			ProcessConnected: false,
		},
	}
}

// Initialize sets up the controller
// 関心の分離: 初期化ロジックの分離
func (c *Controller) Initialize(versions []string) {
	c.config.Versions = versions
	if len(versions) > 0 {
		c.state.SelectedVersion = versions[0]
	}
}

// Start begins the application
// カプセル化: 起動処理の責任
func (c *Controller) Start() error {
	if c.state.Running {
		return fmt.Errorf("already running")
	}

	if c.eventHandler != nil {
		if err := c.eventHandler.OnStart(c.state.SelectedVersion); err != nil {
			return err
		}
	}

	c.state.Running = true
	c.state.StatusText = "Status: Running"
	c.state.MetricsURL = fmt.Sprintf("Metrics: http://localhost:%d/metrics", c.config.DefaultPort)
	
	return nil
}

// Stop halts the application
// カプセル化: 停止処理の責任
func (c *Controller) Stop() error {
	if !c.state.Running {
		return fmt.Errorf("not running")
	}

	if c.eventHandler != nil {
		if err := c.eventHandler.OnStop(); err != nil {
			return err
		}
	}

	c.state.Running = false
	c.state.StatusText = "Status: Stopped"
	c.state.MetricsURL = ""
	// Reset process connection status when stopping
	c.state.ProcessConnected = false
	c.state.ProcessStatus = "Process: Not connected"
	
	return nil
}

// SelectVersion updates the selected version
// 単一責任: バージョン選択の処理
func (c *Controller) SelectVersion(version string) {
	c.state.SelectedVersion = version
	if c.eventHandler != nil {
		c.eventHandler.OnVersionChange(version)
	}
}

// GetState returns the current state
// カプセル化: 状態の読み取りアクセス
func (c *Controller) GetState() *State {
	return c.state
}

// SetEventHandler sets the event handler
// 依存性注入: イベントハンドラーの注入
func (c *Controller) SetEventHandler(handler EventHandler) {
	c.eventHandler = handler
}

// UpdateProcessStatus updates the process connection status
// カプセル化: プロセス状態更新の責任
func (c *Controller) UpdateProcessStatus(connected bool, processName string) {
	c.state.ProcessConnected = connected
	if connected {
		c.state.ProcessStatus = fmt.Sprintf("Process: Connected to %s", processName)
	} else {
		c.state.ProcessStatus = fmt.Sprintf("Process: Waiting for %s...", processName)
	}
}