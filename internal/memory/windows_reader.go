package memory

import (
	"errors"
	"fmt"
	"math"

	"github.com/na2na/yso-exporter/internal/process"
	"golang.org/x/sys/windows"
)

// WindowsMemoryReader は Windows API を使用してプロセスメモリを読み取る実装です。
// カプセル化の原則に従い、Windows固有のメモリアクセスを隠蔽します。
type WindowsMemoryReader struct {
	processHandle windows.Handle
	processID     uint32
	processFinder process.ProcessFinder
}

// NewWindowsMemoryReader は WindowsMemoryReader の新しいインスタンスを作成します。
func NewWindowsMemoryReader(finder process.ProcessFinder) *WindowsMemoryReader {
	return &WindowsMemoryReader{
		processHandle: windows.InvalidHandle,
		processFinder: finder,
	}
}

// OpenProcess は指定されたプロセス名のプロセスハンドルを開きます。
func (r *WindowsMemoryReader) OpenProcess(processName string) error {
	if processName == "" {
		return errors.New("プロセス名が空です")
	}

	// プロセスIDを検索
	processID, err := r.processFinder.FindProcessByName(processName)
	if err != nil {
		return fmt.Errorf("プロセス検索に失敗しました: %w", err)
	}

	// プロセスハンドルを開く
	handle, err := windows.OpenProcess(
		windows.PROCESS_VM_READ|windows.PROCESS_QUERY_INFORMATION,
		false,
		processID,
	)
	if err != nil {
		return fmt.Errorf("プロセスハンドル取得に失敗しました: %w", err)
	}

	r.processHandle = handle
	r.processID = processID
	return nil
}

// Close は開いているプロセスハンドルをクローズします。
func (r *WindowsMemoryReader) Close() error {
	if r.processHandle != windows.InvalidHandle {
		err := windows.CloseHandle(r.processHandle)
		r.processHandle = windows.InvalidHandle
		r.processID = 0
		if err != nil {
			return fmt.Errorf("プロセスハンドルクローズに失敗しました: %w", err)
		}
	}
	return nil
}

// ReadInt32 は指定されたアドレスから32bit整数を読み取ります。
func (r *WindowsMemoryReader) ReadInt32(address uintptr) (int32, error) {
	if err := r.validateReadOperation(address); err != nil {
		return 0, err
	}

	buffer := make([]byte, 4)
	bytesRead, err := r.readProcessMemory(address, buffer)
	if err != nil {
		return 0, fmt.Errorf("メモリ読み取りに失敗しました: %w", err)
	}
	if bytesRead != 4 {
		return 0, fmt.Errorf("読み取りバイト数が不正です: expected=4, actual=%d", bytesRead)
	}

	// リトルエンディアンで int32 に変換
	value := int32(buffer[0]) | int32(buffer[1])<<8 | int32(buffer[2])<<16 | int32(buffer[3])<<24
	return value, nil
}

// ReadFloat32 は指定されたアドレスから32bit浮動小数点数を読み取ります。
func (r *WindowsMemoryReader) ReadFloat32(address uintptr) (float32, error) {
	if err := r.validateReadOperation(address); err != nil {
		return 0, err
	}

	buffer := make([]byte, 4)
	bytesRead, err := r.readProcessMemory(address, buffer)
	if err != nil {
		return 0, fmt.Errorf("メモリ読み取りに失敗しました: %w", err)
	}
	if bytesRead != 4 {
		return 0, fmt.Errorf("読み取りバイト数が不正です: expected=4, actual=%d", bytesRead)
	}

	// リトルエンディアンで uint32 に変換してから float32 に変換
	bits := uint32(buffer[0]) | uint32(buffer[1])<<8 | uint32(buffer[2])<<16 | uint32(buffer[3])<<24
	value := math.Float32frombits(bits)
	return value, nil
}

// ReadByte は指定されたアドレスから1バイトを読み取ります。
func (r *WindowsMemoryReader) ReadByte(address uintptr) (byte, error) {
	if err := r.validateReadOperation(address); err != nil {
		return 0, err
	}

	buffer := make([]byte, 1)
	bytesRead, err := r.readProcessMemory(address, buffer)
	if err != nil {
		return 0, fmt.Errorf("メモリ読み取りに失敗しました: %w", err)
	}
	if bytesRead != 1 {
		return 0, fmt.Errorf("読み取りバイト数が不正です: expected=1, actual=%d", bytesRead)
	}

	return buffer[0], nil
}

// ResolvePointerChain はポインタチェインを辿って最終的なアドレスを解決します。
func (r *WindowsMemoryReader) ResolvePointerChain(baseAddress uintptr, offsets []uintptr) (uintptr, error) {
	if err := r.validateReadOperation(baseAddress); err != nil {
		return 0, err
	}
	if len(offsets) == 0 {
		return 0, errors.New("オフセットスライスが空です")
	}

	currentAddress := baseAddress

	// ポインタチェインを順次辿る
	for i, offset := range offsets {
		// ポインタを読み取り
		pointerValue, err := r.readPointer(currentAddress)
		if err != nil {
			return 0, fmt.Errorf("ポインタ読み取りに失敗しました (step %d): %w", i, err)
		}
		
		// 次のアドレスを計算
		currentAddress = pointerValue + offset
	}

	return currentAddress, nil
}

// validateReadOperation は読み取り操作の前提条件を検証します。
func (r *WindowsMemoryReader) validateReadOperation(address uintptr) error {
	if r.processHandle == windows.InvalidHandle {
		return errors.New("プロセスが開かれていません")
	}
	if address == 0 {
		return errors.New("アドレスがゼロです")
	}
	return nil
}

// readProcessMemory は Windows API を使用してプロセスメモリを読み取ります。
func (r *WindowsMemoryReader) readProcessMemory(address uintptr, buffer []byte) (uintptr, error) {
	var bytesRead uintptr
	err := windows.ReadProcessMemory(
		r.processHandle,
		address,
		&buffer[0],
		uintptr(len(buffer)),
		&bytesRead,
	)
	if err != nil {
		return 0, err
	}
	return bytesRead, nil
}

// readPointer はポインタサイズ（uintptr）の値を読み取ります。
// YSO_WIN.exeは32bitアプリケーションなので、常に32bitポインタとして読み取ります。
func (r *WindowsMemoryReader) readPointer(address uintptr) (uintptr, error) {
	// YSO_WIN.exeは32bitアプリケーションなので、ポインタサイズは4バイト固定
	const pointerSize = 4
	buffer := make([]byte, pointerSize)
	
	bytesRead, err := r.readProcessMemory(address, buffer)
	if err != nil {
		return 0, err
	}
	if bytesRead != pointerSize {
		return 0, fmt.Errorf("ポインタ読み取りバイト数が不正です: expected=%d, actual=%d", pointerSize, bytesRead)
	}

	// 32bitポインタとして値を構築（リトルエンディアン）
	value := uintptr(buffer[0]) | uintptr(buffer[1])<<8 | uintptr(buffer[2])<<16 | uintptr(buffer[3])<<24
	return value, nil
}

// ResolveSumBase は baseA、baseB それぞれのアドレスから32bit値を読み取り、
// 両者を加算した値を返します。
// 用途: LiveSplit ASL の `P = [A] + [B]` のように、2箇所の値の合算でしか
// 求まらないポインタを、既存の「単一ベース + オフセットチェイン」形式では
// 表現できないため、専用の解決手段として提供する。
func (r *WindowsMemoryReader) ResolveSumBase(baseA, baseB uintptr) (uintptr, error) {
	valueA, err := r.ReadInt32(baseA)
	if err != nil {
		return 0, fmt.Errorf("baseAの読み取りに失敗しました: %w", err)
	}

	valueB, err := r.ReadInt32(baseB)
	if err != nil {
		return 0, fmt.Errorf("baseBの読み取りに失敗しました: %w", err)
	}

	sum := uint32(valueA) + uint32(valueB)
	return uintptr(sum), nil
}

// ParseAndResolveAddress はアドレスを解析して実際のメモリアドレスを返します
// 注: この機能は本来WindowsMemoryReaderの責任外だが、
// 既存のインターフェースとの互換性のために保持しています
func (r *WindowsMemoryReader) ParseAndResolveAddress(address interface{}) (uintptr, error) {
	if r.processHandle == 0 {
		return 0, fmt.Errorf("process not open")
	}
	
	// モジュール解決とアドレス解析の責任を専門のコンポーネントに委譲
	resolver := NewWindowsModuleResolver(r.processHandle)
	addr, err := NewAddressFromInterface(address, resolver)
	if err != nil {
		return 0, err
	}
	
	return addr.Value(), nil
}