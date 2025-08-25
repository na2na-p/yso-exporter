package gui

// ApplicationAdapter adapts the main application for GUI use
// カプセル化: アプリケーション適応の単一責任
type ApplicationAdapter interface {
	Start() error
	Stop() error
	IsRunning() bool
	GetProcessName() string
}

// ApplicationLifecycle manages application lifecycle
// 関心の分離: ライフサイクル管理の責任を分離
type ApplicationLifecycle interface {
	Create(version string, port int, interval int) (ApplicationAdapter, error)
	Destroy(adapter ApplicationAdapter) error
}