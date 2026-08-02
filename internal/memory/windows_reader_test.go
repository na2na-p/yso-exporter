package memory_test

import (
	"testing"
	"unsafe"

	"github.com/google/go-cmp/cmp"
	"github.com/na2na/yso-exporter/internal/memory"
	"github.com/na2na/yso-exporter/internal/process"
)

func TestWindowsMemoryReader_OpenProcess(t *testing.T) {
	type args struct {
		processName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "正常系: 存在しないプロセス名でエラーが返る（テスト環境での期待動作）",
			args: args{
				processName: "nonexistent_process.exe",
			},
			wantErr: true,
		},
		{
			name: "異常系: 空のプロセス名でエラーが返る",
			args: args{
				processName: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ProcessFinderを実際の実装で使用
			finder := process.NewWindowsProcessFinder()
			reader := memory.NewWindowsMemoryReader(finder)
			err := reader.OpenProcess(tt.args.processName)
			if (err != nil) != tt.wantErr {
				t.Errorf("WindowsMemoryReader.OpenProcess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWindowsMemoryReader_Close(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "正常系: プロセスハンドルが未開時でもエラーにならない",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := process.NewWindowsProcessFinder()
		reader := memory.NewWindowsMemoryReader(finder)
			if err := reader.Close(); (err != nil) != tt.wantErr {
				t.Errorf("WindowsMemoryReader.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWindowsMemoryReader_ReadInt32(t *testing.T) {
	type args struct {
		address uintptr
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		{
			name: "異常系: プロセス未オープン時にエラーが返る",
			args: args{
				address: 0x1000,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "異常系: ゼロアドレス時にエラーが返る",
			args: args{
				address: 0,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := process.NewWindowsProcessFinder()
		reader := memory.NewWindowsMemoryReader(finder)
			got, err := reader.ReadInt32(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("WindowsMemoryReader.ReadInt32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("WindowsMemoryReader.ReadInt32() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWindowsMemoryReader_ReadFloat32(t *testing.T) {
	type args struct {
		address uintptr
	}
	tests := []struct {
		name    string
		args    args
		want    float32
		wantErr bool
	}{
		{
			name: "異常系: プロセス未オープン時にエラーが返る",
			args: args{
				address: 0x1000,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "異常系: ゼロアドレス時にエラーが返る",
			args: args{
				address: 0,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := process.NewWindowsProcessFinder()
		reader := memory.NewWindowsMemoryReader(finder)
			got, err := reader.ReadFloat32(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("WindowsMemoryReader.ReadFloat32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("WindowsMemoryReader.ReadFloat32() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWindowsMemoryReader_ReadByte(t *testing.T) {
	type args struct {
		address uintptr
	}
	tests := []struct {
		name    string
		args    args
		want    byte
		wantErr bool
	}{
		{
			name: "異常系: プロセス未オープン時にエラーが返る",
			args: args{
				address: 0x1000,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "異常系: ゼロアドレス時にエラーが返る",
			args: args{
				address: 0,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := process.NewWindowsProcessFinder()
		reader := memory.NewWindowsMemoryReader(finder)
			got, err := reader.ReadByte(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("WindowsMemoryReader.ReadByte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("WindowsMemoryReader.ReadByte() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWindowsMemoryReader_ResolvePointerChain(t *testing.T) {
	type args struct {
		baseAddress uintptr
		offsets     []uintptr
	}
	tests := []struct {
		name    string
		args    args
		want    uintptr
		wantErr bool
	}{
		{
			name: "異常系: プロセス未オープン時にエラーが返る",
			args: args{
				baseAddress: 0x1000,
				offsets:     []uintptr{0x10, 0x20},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "異常系: ベースアドレスがゼロ時にエラーが返る",
			args: args{
				baseAddress: 0,
				offsets:     []uintptr{0x10},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "異常系: オフセットスライスが空時にエラーが返る",
			args: args{
				baseAddress: 0x1000,
				offsets:     []uintptr{},
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := process.NewWindowsProcessFinder()
		reader := memory.NewWindowsMemoryReader(finder)
			got, err := reader.ResolvePointerChain(tt.args.baseAddress, tt.args.offsets)
			if (err != nil) != tt.wantErr {
				t.Errorf("WindowsMemoryReader.ResolvePointerChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("WindowsMemoryReader.ResolvePointerChain() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWindowsMemoryReader_ResolveSumBase(t *testing.T) {
	type args struct {
		baseA uintptr
		baseB uintptr
	}
	tests := []struct {
		name    string
		args    args
		want    uintptr
		wantErr bool
	}{
		{
			name: "異常系: プロセス未オープン時にエラーが返る",
			args: args{
				baseA: 0x1000,
				baseB: 0x2000,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "異常系: baseAがゼロ時にエラーが返る",
			args: args{
				baseA: 0,
				baseB: 0x2000,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "異常系: baseBがゼロ時にエラーが返る",
			args: args{
				baseA: 0x1000,
				baseB: 0,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := process.NewWindowsProcessFinder()
			reader := memory.NewWindowsMemoryReader(finder)
			got, err := reader.ResolveSumBase(tt.args.baseA, tt.args.baseB)
			if (err != nil) != tt.wantErr {
				t.Errorf("WindowsMemoryReader.ResolveSumBase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("WindowsMemoryReader.ResolveSumBase() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// テスト用の補助関数
func createTestData() []byte {
	var data int32 = 12345
	return (*[4]byte)(unsafe.Pointer(&data))[:]
}