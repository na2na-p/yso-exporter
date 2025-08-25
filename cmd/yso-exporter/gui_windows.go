//go:build windows

package main

import (
	"log"

	"github.com/lxn/walk"
	"github.com/na2na/yso-exporter/internal/debug"
	"github.com/na2na/yso-exporter/internal/gui"
)

// GUIApplication handles the GUI lifecycle
// カプセル化: GUI アプリケーションの単一責任
type GUIApplication struct {
	controller   *gui.Controller
	coordinator  *gui.EventCoordinator
	mainWindow   *walk.MainWindow
	builder      *gui.WindowBuilder
}

// runGUI is the entry point for GUI mode
// 関心の分離: エントリーポイントの責任のみ
func runGUI() {
	defer debug.LogPanic("runGUI")
	
	debug.Log("Creating GUI application")
	guiApp := &GUIApplication{}
	
	debug.Log("Starting GUI application")
	if err := guiApp.Run(); err != nil {
		debug.LogError(err, "GUI Run failed")
		log.Fatal(err)
	}
	debug.Log("GUI application terminated normally")
}

// Run starts the GUI application
// カプセル化: GUI実行の責任
func (g *GUIApplication) Run() error {
	defer debug.LogPanic("GUIApplication.Run")
	trace := debug.Enter("GUIApplication.Run")
	defer trace.Exit()
	
	// Initialize configuration
	debug.Log("Initializing GUI configuration")
	guiConfig := gui.NewGUIConfig()
	versions := getAvailableVersions()
	debug.Log("Available versions: %v", versions)
	
	// Create controller
	debug.Log("Creating controller")
	g.controller = gui.NewController(guiConfig)
	g.controller.Initialize(versions)
	debug.Log("Controller initialized with %d versions", len(versions))
	
	// Create application lifecycle with process callback
	debug.Log("Creating application lifecycle")
	processCallback := &guiProcessCallback{app: g}
	lifecycle := NewApplicationLifecycle(processCallback)
	debug.Log("Application lifecycle created")
	
	// Create event coordinator
	debug.Log("Creating event coordinator")
	g.coordinator = gui.NewEventCoordinator(g.controller, lifecycle)
	g.controller.SetEventHandler(g.coordinator)
	debug.Log("Event coordinator set")
	
	// Build window
	debug.Log("Building main window")
	if err := g.buildWindow(); err != nil {
		debug.LogError(err, "Failed to build window")
		return err
	}
	debug.Log("Main window built successfully")
	
	// No need for process notifier - coordinator handles it directly
	debug.Log("Event coordinator configured")
	
	// Setup system tray
	debug.Log("Setting up system tray")
	if err := g.setupSystemTray(); err != nil {
		debug.LogError(err, "Failed to setup system tray")
		return err
	}
	debug.Log("System tray setup complete")
	
	// Run the main window
	debug.Log("Starting main window event loop")
	g.mainWindow.Run()
	debug.Log("Main window event loop terminated")
	return nil
}

// buildWindow builds the main window
// 単一責任: ウィンドウ構築の責任
func (g *GUIApplication) buildWindow() error {
	defer debug.LogPanic("buildWindow")
	trace := debug.Enter("buildWindow")
	defer trace.Exit()
	
	var versionCombo *walk.ComboBox
	var statusLabel, processLabel, urlLabel *walk.Label
	
	debug.Log("Creating window builder")
	g.builder = gui.NewWindowBuilder(g.controller)
	g.builder.SetMainWindow(&g.mainWindow)
	g.builder.SetControls(&versionCombo, &statusLabel, &processLabel, &urlLabel)
	debug.Log("Window builder configured")
	
	debug.Log("Building window definition")
	windowDef := g.builder.Build()
	
	debug.Log("Creating window from definition")
	if err := windowDef.Create(); err != nil {
		debug.LogError(err, "Window creation failed")
		return err
	}
	debug.Log("Window created successfully")
	
	// Set initial version selection
	if versionCombo != nil && len(g.controller.GetState().SelectedVersion) > 0 {
		versionCombo.SetCurrentIndex(0)
	}
	
	return nil
}

// setupSystemTray sets up the system tray icon
// 単一責任: システムトレイ設定の責任
func (g *GUIApplication) setupSystemTray() error {
	ni, err := walk.NewNotifyIcon(g.mainWindow)
	if err != nil {
		return err
	}
	
	ni.SetToolTip("YSO Exporter")
	ni.SetVisible(true)
	
	// Minimize to tray
	g.mainWindow.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		*canceled = true
		g.mainWindow.Hide()
	})
	
	// Restore from tray
	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			g.mainWindow.Show()
		}
	})
	
	// Exit action
	exitAction := walk.NewAction()
	exitAction.SetText("Exit")
	exitAction.Triggered().Attach(func() {
		g.coordinator.OnStop()
		walk.App().Exit(0)
	})
	
	ni.ContextMenu().Actions().Add(exitAction)
	
	return nil
}

// guiProcessCallback handles process status callbacks
// カプセル化: プロセス状態コールバックの実装
type guiProcessCallback struct {
	app *GUIApplication
}

// OnProcessConnected handles process connection
// 単一責任: プロセス接続通知の処理
func (c *guiProcessCallback) OnProcessConnected(processName string) {
	debug.Log("Process connected callback: %s", processName)
	
	if c.app == nil || c.app.mainWindow == nil {
		debug.Log("WARNING: app or mainWindow is nil, cannot process connection")
		return
	}
	
	// UIスレッドで安全に実行
	debug.Log("Synchronizing UI thread for process connection")
	c.app.mainWindow.Synchronize(func() {
		defer debug.LogPanic("OnProcessConnected.Synchronize")
		
		if c.app.coordinator != nil {
			debug.Log("Notifying coordinator of process connection")
			c.app.coordinator.OnProcessConnected(processName)
			// Update UI
			if c.app.builder != nil {
				c.app.builder.UpdateProcessStatus()
			}
		} else {
			debug.Log("WARNING: coordinator is nil")
		}
	})
	debug.Log("Process connection handled")
}

// OnProcessLost handles process disconnection
// 単一責任: プロセス切断通知の処理
func (c *guiProcessCallback) OnProcessLost(processName string) {
	debug.Log("Process lost callback: %s", processName)
	
	if c.app == nil || c.app.mainWindow == nil {
		debug.Log("WARNING: app or mainWindow is nil, cannot process disconnection")
		return
	}
	
	// UIスレッドで安全に実行
	debug.Log("Synchronizing UI thread for process disconnection")
	c.app.mainWindow.Synchronize(func() {
		defer debug.LogPanic("OnProcessLost.Synchronize")
		
		if c.app.coordinator != nil {
			debug.Log("Notifying coordinator of process loss")
			c.app.coordinator.OnProcessLost(processName)
			// Update UI
			if c.app.builder != nil {
				c.app.builder.UpdateProcessStatus()
			}
		} else {
			debug.Log("WARNING: coordinator is nil")
		}
	})
	debug.Log("Process disconnection handled")
}