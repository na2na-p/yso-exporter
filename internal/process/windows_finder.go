package process

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WindowsProcessFinder はWindows環境でプロセスを検索する実装です
type WindowsProcessFinder struct{}

// NewWindowsProcessFinder は新しいWindowsProcessFinderを作成します
func NewWindowsProcessFinder() *WindowsProcessFinder {
	return &WindowsProcessFinder{}
}

// FindProcessByName は指定された名前のプロセスIDを検索します
func (f *WindowsProcessFinder) FindProcessByName(processName string) (uint32, error) {
	// プロセススナップショットを作成
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, fmt.Errorf("プロセススナップショット作成に失敗しました: %w", err)
	}
	defer windows.CloseHandle(snapshot)

	// プロセスエントリ構造体
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))

	// 最初のプロセスを取得
	err = windows.Process32First(snapshot, &pe)
	if err != nil {
		return 0, fmt.Errorf("最初のプロセス取得に失敗しました: %w", err)
	}

	// プロセス名を比較して検索
	for {
		exeFile := syscall.UTF16ToString(pe.ExeFile[:])
		if exeFile == processName {
			return pe.ProcessID, nil
		}

		// 次のプロセスを取得
		err = windows.Process32Next(snapshot, &pe)
		if err != nil {
			break
		}
	}

	return 0, fmt.Errorf("%w: '%s'", ErrProcessNotFound, processName)
}