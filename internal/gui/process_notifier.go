package gui

// ProcessNotifier notifies GUI about process status changes
// カプセル化: プロセス状態通知の単一責任
type ProcessNotifier struct {
	controller *Controller
	builder    *WindowBuilder
	mainWindow interface{} // walk.MainWindow interface
}

// NewProcessNotifier creates a new process notifier
// 単一責任: 通知機能の生成
func NewProcessNotifier(controller *Controller, builder *WindowBuilder) *ProcessNotifier {
	return &ProcessNotifier{
		controller: controller,
		builder:    builder,
	}
}

// SetMainWindow sets the main window for synchronization
// 依存性注入: ウィンドウ参照の注入
func (n *ProcessNotifier) SetMainWindow(window interface{}) {
	n.mainWindow = window
}

// NotifyProcessConnected notifies that process is connected
// カプセル化: 接続通知の責任
func (n *ProcessNotifier) NotifyProcessConnected(processName string) {
	if n.controller != nil {
		n.controller.UpdateProcessStatus(true, processName)
	}
	n.updateUI()
}

// NotifyProcessLost notifies that process is lost
// カプセル化: 切断通知の責任
func (n *ProcessNotifier) NotifyProcessLost(processName string) {
	if n.controller != nil {
		n.controller.UpdateProcessStatus(false, processName)
	}
	n.updateUI()
}

// updateUI updates the UI to reflect current state
// 関心の分離: UI更新ロジックの分離
func (n *ProcessNotifier) updateUI() {
	if n.builder != nil {
		n.builder.UpdateProcessStatus()
	}
}