package memory

//go:generate mockgen -source=reader.go -destination=mocks/mock_reader.go -package=mocks

// MemoryReader は、プロセスメモリを読み取るためのインターフェースです。
// カプセル化の原則に従い、プロセスメモリアクセスという単一の目的を持ちます。
type MemoryReader interface {
	// OpenProcess は指定されたプロセス名のプロセスハンドルを開きます
	OpenProcess(processName string) error
	
	// Close は開いているプロセスハンドルをクローズします
	Close() error
	
	// ReadInt32 は指定されたアドレスから32bit整数を読み取ります
	ReadInt32(address uintptr) (int32, error)
	
	// ReadFloat32 は指定されたアドレスから32bit浮動小数点数を読み取ります
	ReadFloat32(address uintptr) (float32, error)
	
	// ReadByte は指定されたアドレスから1バイトを読み取ります
	ReadByte(address uintptr) (byte, error)
	
	// ResolvePointerChain はポインタチェインを辿って最終的なアドレスを解決します
	ResolvePointerChain(baseAddress uintptr, offsets []uintptr) (uintptr, error)

	// ParseAndResolveAddress はアドレス文字列を解析して実際のメモリアドレスを返します
	ParseAndResolveAddress(address interface{}) (uintptr, error)

	// ResolveSumBase は baseA、baseB それぞれのアドレスから32bit値を読み取り、
	// 両者を加算した値を返します（2箇所のアドレスの合算で求まるポインタ表現の解決用）
	ResolveSumBase(baseA, baseB uintptr) (uintptr, error)
}