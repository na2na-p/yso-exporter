//go:build windows

package gui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// WindowBuilder builds the GUI window
// カプセル化: UI構築の単一責任
type WindowBuilder struct {
	controller      *Controller
	mainWindow      **walk.MainWindow
	versionCombo    **walk.ComboBox
	statusLabel     **walk.Label
	urlLabel        **walk.Label
	processLabel    **walk.Label
}

// NewWindowBuilder creates a new window builder
// 単一責任: ビルダーの生成のみ
func NewWindowBuilder(controller *Controller) *WindowBuilder {
	return &WindowBuilder{
		controller: controller,
	}
}

// Build creates the main window
// 関心の分離: UI定義とロジックの分離
func (wb *WindowBuilder) Build() MainWindow {
	return MainWindow{
		AssignTo: wb.mainWindow,
		Title:    "YSO Exporter",
		MinSize:  Size{400, 200},
		Size:     Size{400, 200},
		Layout:   VBox{},
		Children: wb.createChildren(),
	}
}

// createChildren creates the window children widgets
// カプセル化: ウィジェット作成の責任
func (wb *WindowBuilder) createChildren() []Widget {
	return []Widget{
		wb.createVersionSelector(),
		wb.createStatusLabel(),
		wb.createProcessLabel(),
		wb.createURLLabel(),
		HSpacer{},
		wb.createControlButtons(),
	}
}

// createVersionSelector creates the version selection composite
// 単一責任: バージョン選択UIの作成
func (wb *WindowBuilder) createVersionSelector() Widget {
	return Composite{
		Layout: HBox{},
		Children: []Widget{
			Label{Text: "Version:"},
			ComboBox{
				AssignTo: wb.versionCombo,
				Model:    wb.controller.config.Versions,
				OnCurrentIndexChanged: func() {
					if combo := *wb.versionCombo; combo != nil && combo.CurrentIndex() >= 0 {
						version := wb.controller.config.Versions[combo.CurrentIndex()]
						wb.controller.SelectVersion(version)
					}
				},
			},
		},
	}
}

// createStatusLabel creates the status label
// 単一責任: ステータス表示の作成
func (wb *WindowBuilder) createStatusLabel() Widget {
	return Label{
		AssignTo: wb.statusLabel,
		Text:     "Status: Stopped",
	}
}

// createProcessLabel creates the process status label
// 単一責任: プロセス状態表示の作成
func (wb *WindowBuilder) createProcessLabel() Widget {
	return Label{
		AssignTo: wb.processLabel,
		Text:     "Process: Not connected",
	}
}

// createURLLabel creates the URL label
// 単一責任: URL表示の作成
func (wb *WindowBuilder) createURLLabel() Widget {
	return Label{
		AssignTo: wb.urlLabel,
		Text:     "",
	}
}

// createControlButtons creates the control buttons composite
// 単一責任: 制御ボタンの作成
func (wb *WindowBuilder) createControlButtons() Widget {
	return Composite{
		Layout: HBox{},
		Children: []Widget{
			wb.createStartButton(),
			wb.createStopButton(),
		},
	}
}

// createStartButton creates the start button
// カプセル化: 開始ボタンの作成とイベント処理
func (wb *WindowBuilder) createStartButton() Widget {
	return PushButton{
		Text: "Start",
		OnClicked: func() {
			if err := wb.controller.Start(); err == nil {
				wb.updateUI()
			}
		},
	}
}

// createStopButton creates the stop button
// カプセル化: 停止ボタンの作成とイベント処理
func (wb *WindowBuilder) createStopButton() Widget {
	return PushButton{
		Text: "Stop",
		OnClicked: func() {
			if err := wb.controller.Stop(); err == nil {
				wb.updateUI()
			}
		},
	}
}

// updateUI updates the UI based on current state
// 関心の分離: UI更新ロジックの分離
func (wb *WindowBuilder) updateUI() {
	state := wb.controller.GetState()
	
	if *wb.statusLabel != nil {
		(*wb.statusLabel).SetText(state.StatusText)
	}
	
	if *wb.processLabel != nil {
		(*wb.processLabel).SetText(state.ProcessStatus)
	}
	
	if *wb.urlLabel != nil {
		(*wb.urlLabel).SetText(state.MetricsURL)
	}
	
	// Disable version selector when running or process is connected
	if *wb.versionCombo != nil {
		(*wb.versionCombo).SetEnabled(!state.Running && !state.ProcessConnected)
	}
}

// UpdateProcessStatus updates the process status label
// 単一責任: プロセス状態表示の更新
func (wb *WindowBuilder) UpdateProcessStatus() {
	if wb.processLabel != nil && *wb.processLabel != nil {
		state := wb.controller.GetState()
		(*wb.processLabel).SetText(state.ProcessStatus)
		
		// Also update version selector enabled state
		if wb.versionCombo != nil && *wb.versionCombo != nil {
			(*wb.versionCombo).SetEnabled(!state.Running && !state.ProcessConnected)
		}
	}
}

// SetMainWindow sets the main window reference
// 依存性注入: ウィンドウ参照の注入
func (wb *WindowBuilder) SetMainWindow(window **walk.MainWindow) {
	wb.mainWindow = window
}

// SetControls sets the control references
// 依存性注入: コントロール参照の注入
func (wb *WindowBuilder) SetControls(versionCombo **walk.ComboBox, statusLabel, processLabel, urlLabel **walk.Label) {
	wb.versionCombo = versionCombo
	wb.statusLabel = statusLabel
	wb.processLabel = processLabel
	wb.urlLabel = urlLabel
}