package process

import "fmt"

// ProcessFinder はプロセスを検索するためのインターフェースです
type ProcessFinder interface {
	// FindProcessByName は指定された名前のプロセスIDを検索します
	FindProcessByName(processName string) (uint32, error)
}

// ErrProcessNotFound はプロセスが見つからなかった場合のエラーです
var ErrProcessNotFound = fmt.Errorf("プロセスが見つかりませんでした")