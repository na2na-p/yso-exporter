package memory

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ModuleResolver はプロセス内のモジュール情報を解決する責任を持ちます
// 単一の目的: モジュールのベースアドレスの解決
type ModuleResolver interface {
	GetModuleBase(moduleName string) (uintptr, error)
}

// WindowsModuleResolver はWindows APIを使用してモジュール情報を解決します
// カプセル化: モジュール解決に必要な全ての知識をカプセル化
type WindowsModuleResolver struct {
	processHandle windows.Handle
}

// NewWindowsModuleResolver は新しいWindowsModuleResolverを作成します
func NewWindowsModuleResolver(processHandle windows.Handle) *WindowsModuleResolver {
	return &WindowsModuleResolver{
		processHandle: processHandle,
	}
}

// GetModuleBase は指定されたモジュールのベースアドレスを取得します
// 単一の責任: モジュールベースアドレスの取得のみ
func (r *WindowsModuleResolver) GetModuleBase(moduleName string) (uintptr, error) {
	if r.processHandle == 0 {
		return 0, fmt.Errorf("invalid process handle")
	}
	
	snapshot, err := r.createModuleSnapshot()
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot)
	
	return r.findModuleBase(snapshot, moduleName)
}

// createModuleSnapshot はモジュールのスナップショットを作成します
// 単一の責任: スナップショットの作成
func (r *WindowsModuleResolver) createModuleSnapshot() (windows.Handle, error) {
	processID, err := windows.GetProcessId(r.processHandle)
	if err != nil {
		return 0, fmt.Errorf("failed to get process ID: %w", err)
	}
	
	snapshot, err := windows.CreateToolhelp32Snapshot(
		windows.TH32CS_SNAPMODULE|windows.TH32CS_SNAPMODULE32,
		processID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create module snapshot: %w", err)
	}
	
	return snapshot, nil
}

// findModuleBase はスナップショットから指定モジュールのベースアドレスを検索します
// 単一の責任: モジュールの検索
func (r *WindowsModuleResolver) findModuleBase(snapshot windows.Handle, moduleName string) (uintptr, error) {
	iterator := NewModuleIterator(snapshot)
	targetModule := strings.ToUpper(moduleName)
	
	for iterator.Next() {
		entry := iterator.Current()
		currentModule := syscall.UTF16ToString(entry.Module[:])
		
		if strings.ToUpper(currentModule) == targetModule {
			return uintptr(entry.ModBaseAddr), nil
		}
	}
	
	if err := iterator.Err(); err != nil {
		return 0, fmt.Errorf("module iteration failed: %w", err)
	}
	
	return 0, fmt.Errorf("module %s not found in process", moduleName)
}

// ModuleIterator はモジュールエントリのイテレーションを責任とします
// 単一の目的: モジュールエントリの順次アクセス
type ModuleIterator struct {
	snapshot windows.Handle
	current  windows.ModuleEntry32
	err      error
	first    bool
}

// NewModuleIterator は新しいModuleIteratorを作成します
func NewModuleIterator(snapshot windows.Handle) *ModuleIterator {
	return &ModuleIterator{
		snapshot: snapshot,
		first:    true,
	}
}

// Next は次のモジュールエントリに移動します
func (it *ModuleIterator) Next() bool {
	if it.err != nil {
		return false
	}
	
	if it.first {
		it.current.Size = uint32(unsafe.Sizeof(it.current))
		it.err = windows.Module32First(it.snapshot, &it.current)
		it.first = false
	} else {
		it.err = windows.Module32Next(it.snapshot, &it.current)
	}
	
	if it.err == syscall.ERROR_NO_MORE_FILES {
		it.err = nil // This is not an error, just end of iteration
		return false
	}
	
	return it.err == nil
}

// Current は現在のモジュールエントリを返します
func (it *ModuleIterator) Current() windows.ModuleEntry32 {
	return it.current
}

// Err はイテレーション中に発生したエラーを返します
func (it *ModuleIterator) Err() error {
	return it.err
}