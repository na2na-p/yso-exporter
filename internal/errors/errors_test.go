package errors_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	yerrors "github.com/na2na/yso-exporter/internal/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestErrorCode_String(t *testing.T) {
	tests := []struct {
		name string
		code yerrors.ErrorCode
		want string
	}{
		{
			name: "正常系: NotFoundコードの文字列表現",
			code: yerrors.CodeNotFound,
			want: "NOT_FOUND",
		},
		{
			name: "正常系: Permissionコードの文字列表現",
			code: yerrors.CodePermission,
			want: "PERMISSION",
		},
		{
			name: "正常系: Configコードの文字列表現",
			code: yerrors.CodeConfig,
			want: "CONFIG",
		},
		{
			name: "正常系: IOコードの文字列表現",
			code: yerrors.CodeIO,
			want: "IO",
		},
		{
			name: "正常系: Internalコードの文字列表現",
			code: yerrors.CodeInternal,
			want: "INTERNAL",
		},
		{
			name: "正常系: ProcessNotFoundコードの文字列表現",
			code: yerrors.CodeProcessNotFound,
			want: "PROCESS_NOT_FOUND",
		},
		{
			name: "正常系: MemoryReadコードの文字列表現",
			code: yerrors.CodeMemoryRead,
			want: "MEMORY_READ",
		},
		{
			name: "正常系: Unknownコードの文字列表現",
			code: yerrors.CodeUnknown,
			want: "UNKNOWN",
		},
		{
			name: "境界値: 不正なコード値",
			code: yerrors.ErrorCode(999),
			want: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.code.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("String() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestYSOError_Error(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name     string
		ysoError *yerrors.YSOError
		want     string
	}{
		{
			name:     "正常系: ラップされていないエラー",
			ysoError: yerrors.NewError(yerrors.CodeNotFound, "resource not found"),
			want:     "NOT_FOUND: resource not found",
		},
		{
			name:     "正常系: ラップされたエラー",
			ysoError: yerrors.WrapError(yerrors.CodeIO, "read failed", originalErr),
			want:     "IO: read failed: original error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ysoError.Error()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Error() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestYSOError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name     string
		ysoError *yerrors.YSOError
		want     error
	}{
		{
			name:     "正常系: ラップされていないエラー",
			ysoError: yerrors.NewError(yerrors.CodeNotFound, "resource not found"),
			want:     nil,
		},
		{
			name:     "正常系: ラップされたエラー",
			ysoError: yerrors.WrapError(yerrors.CodeIO, "read failed", originalErr),
			want:     originalErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ysoError.Unwrap()
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Unwrap() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestYSOError_Code(t *testing.T) {
	tests := []struct {
		name     string
		ysoError *yerrors.YSOError
		want     yerrors.ErrorCode
	}{
		{
			name:     "正常系: NotFoundコード",
			ysoError: yerrors.NewError(yerrors.CodeNotFound, "not found"),
			want:     yerrors.CodeNotFound,
		},
		{
			name:     "正常系: MemoryReadコード",
			ysoError: yerrors.NewError(yerrors.CodeMemoryRead, "memory read failed"),
			want:     yerrors.CodeMemoryRead,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ysoError.Code()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Code() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestYSOError_IsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		ysoError *yerrors.YSOError
		want     bool
	}{
		{
			name:     "正常系: リトライ可能なエラー (MemoryRead)",
			ysoError: yerrors.NewError(yerrors.CodeMemoryRead, "memory read failed"),
			want:     true,
		},
		{
			name:     "正常系: リトライ可能なエラー (ProcessNotFound)",
			ysoError: yerrors.NewError(yerrors.CodeProcessNotFound, "process not found"),
			want:     true,
		},
		{
			name:     "正常系: リトライ不可能なエラー (Config)",
			ysoError: yerrors.NewError(yerrors.CodeConfig, "config error"),
			want:     false,
		},
		{
			name:     "正常系: リトライ不可能なエラー (Permission)",
			ysoError: yerrors.NewError(yerrors.CodePermission, "permission denied"),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ysoError.IsRetryable()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("IsRetryable() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewProcessNotFoundError(t *testing.T) {
	tests := []struct {
		name        string
		processName string
		wantCode    yerrors.ErrorCode
		wantMessage string
	}{
		{
			name:        "正常系: プロセス名を指定したエラー作成",
			processName: "YS_ORIGIN.exe",
			wantCode:    yerrors.CodeProcessNotFound,
			wantMessage: "PROCESS_NOT_FOUND: process not found: YS_ORIGIN.exe",
		},
		{
			name:        "境界値: 空のプロセス名",
			processName: "",
			wantCode:    yerrors.CodeProcessNotFound,
			wantMessage: "PROCESS_NOT_FOUND: process not found: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := yerrors.NewProcessNotFoundError(tt.processName)
			
			if diff := cmp.Diff(tt.wantCode, err.Code()); diff != "" {
				t.Errorf("Code() mismatch (-want +got):\n%s", diff)
			}
			
			if diff := cmp.Diff(tt.wantMessage, err.Error()); diff != "" {
				t.Errorf("Error() mismatch (-want +got):\n%s", diff)
			}
			
			if !err.IsRetryable() {
				t.Error("ProcessNotFoundError should be retryable")
			}
		})
	}
}

func TestNewMemoryReadError(t *testing.T) {
	originalErr := errors.New("access denied")

	tests := []struct {
		name        string
		address     uintptr
		err         error
		wantCode    yerrors.ErrorCode
		wantMessage string
		wantWrapped error
	}{
		{
			name:        "正常系: メモリ読み取りエラー作成",
			address:     0x12345678,
			err:         originalErr,
			wantCode:    yerrors.CodeMemoryRead,
			wantMessage: "MEMORY_READ: failed to read memory at address 0x12345678: access denied",
			wantWrapped: originalErr,
		},
		{
			name:        "境界値: アドレス0",
			address:     0x0,
			err:         originalErr,
			wantCode:    yerrors.CodeMemoryRead,
			wantMessage: "MEMORY_READ: failed to read memory at address 0x0: access denied",
			wantWrapped: originalErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := yerrors.NewMemoryReadError(tt.address, tt.err)
			
			if diff := cmp.Diff(tt.wantCode, err.Code()); diff != "" {
				t.Errorf("Code() mismatch (-want +got):\n%s", diff)
			}
			
			if diff := cmp.Diff(tt.wantMessage, err.Error()); diff != "" {
				t.Errorf("Error() mismatch (-want +got):\n%s", diff)
			}
			
			if diff := cmp.Diff(tt.wantWrapped, err.Unwrap(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Unwrap() mismatch (-want +got):\n%s", diff)
			}
			
			if !err.IsRetryable() {
				t.Error("MemoryReadError should be retryable")
			}
		})
	}
}

func TestNewConfigLoadError(t *testing.T) {
	originalErr := errors.New("file not found")

	tests := []struct {
		name        string
		path        string
		err         error
		wantCode    yerrors.ErrorCode
		wantMessage string
		wantWrapped error
	}{
		{
			name:        "正常系: 設定ファイル読み込みエラー作成",
			path:        "/path/to/config.yaml",
			err:         originalErr,
			wantCode:    yerrors.CodeConfig,
			wantMessage: "CONFIG: failed to load config from /path/to/config.yaml: file not found",
			wantWrapped: originalErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := yerrors.NewConfigLoadError(tt.path, tt.err)
			
			if diff := cmp.Diff(tt.wantCode, err.Code()); diff != "" {
				t.Errorf("Code() mismatch (-want +got):\n%s", diff)
			}
			
			if diff := cmp.Diff(tt.wantMessage, err.Error()); diff != "" {
				t.Errorf("Error() mismatch (-want +got):\n%s", diff)
			}
			
			if diff := cmp.Diff(tt.wantWrapped, err.Unwrap(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Unwrap() mismatch (-want +got):\n%s", diff)
			}
			
			if err.IsRetryable() {
				t.Error("ConfigLoadError should not be retryable")
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	retryableErr := yerrors.NewError(yerrors.CodeMemoryRead, "memory read failed")
	nonRetryableErr := yerrors.NewError(yerrors.CodeConfig, "config error")
	standardErr := errors.New("standard error")

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "正常系: リトライ可能なYSOError",
			err:  retryableErr,
			want: true,
		},
		{
			name: "正常系: リトライ不可能なYSOError",
			err:  nonRetryableErr,
			want: false,
		},
		{
			name: "正常系: 標準エラー",
			err:  standardErr,
			want: false,
		},
		{
			name: "境界値: nilエラー",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := yerrors.IsRetryable(tt.err)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("IsRetryable() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRetry(t *testing.T) {
	tests := []struct {
		name        string
		policy      *yerrors.RetryPolicy
		fn          yerrors.RetryFunc
		wantErr     bool
		wantAttempts int
	}{
		{
			name:   "正常系: 成功する関数",
			policy: yerrors.DefaultRetryPolicy(),
			fn: func(ctx context.Context) error {
				return nil
			},
			wantErr:      false,
			wantAttempts: 1,
		},
		{
			name:   "正常系: リトライ不可能なエラー",
			policy: yerrors.DefaultRetryPolicy(),
			fn: func(ctx context.Context) error {
				return yerrors.NewError(yerrors.CodeConfig, "config error")
			},
			wantErr:      true,
			wantAttempts: 1,
		},
		{
			name:   "正常系: リトライ可能なエラーで最大回数まで失敗",
			policy: &yerrors.RetryPolicy{MaxRetries: 2, BaseDelay: 1 * time.Millisecond, RetryableBy: yerrors.IsRetryable},
			fn: func(ctx context.Context) error {
				return yerrors.NewError(yerrors.CodeMemoryRead, "memory read failed")
			},
			wantErr:      true,
			wantAttempts: 3, // 初回 + 2回リトライ
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			attempts := 0
			
			wrappedFn := func(ctx context.Context) error {
				attempts++
				return tt.fn(ctx)
			}
			
			err := yerrors.Retry(ctx, tt.policy, wrappedFn)
			
			if tt.wantErr && err == nil {
				t.Error("Retry() error = nil, wantErr true")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Retry() unexpected error: %v", err)
			}
			
			if diff := cmp.Diff(tt.wantAttempts, attempts); diff != "" {
				t.Errorf("attempts mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // すぐにキャンセル
	
	policy := &yerrors.RetryPolicy{
		MaxRetries:  3,
		BaseDelay:   100 * time.Millisecond,
		RetryableBy: yerrors.IsRetryable,
	}
	
	fn := func(ctx context.Context) error {
		return yerrors.NewError(yerrors.CodeMemoryRead, "memory read failed")
	}
	
	err := yerrors.Retry(ctx, policy, fn)
	
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestWithRecover(t *testing.T) {
	tests := []struct {
		name          string
		fn            func() error
		recoverFn     yerrors.RecoverFunc
		expectPanic   bool
		expectRecover bool
		wantErr       error
	}{
		{
			name: "正常系: パニックなしの正常実行",
			fn: func() error {
				return nil
			},
			recoverFn:     nil,
			expectPanic:   false,
			expectRecover: false,
			wantErr:       nil,
		},
		{
			name: "正常系: パニックありでリカバリー関数実行",
			fn: func() error {
				panic("test panic")
			},
			recoverFn: func(r interface{}) error {
				if r.(string) != "test panic" {
					t.Errorf("expected 'test panic', got %v", r)
				}
				return fmt.Errorf("recovered: %v", r)
			},
			expectPanic:   true,
			expectRecover: true,
			wantErr:       nil, // WithRecoverはリカバリー関数の戻り値を返さない
		},
		{
			name: "正常系: パニックありでリカバリー関数なし",
			fn: func() error {
				panic("test panic")
			},
			recoverFn:     nil,
			expectPanic:   true,
			expectRecover: false,
			wantErr:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recovered := false
			recoverFn := tt.recoverFn
			if recoverFn != nil {
				originalFn := recoverFn
				recoverFn = func(r interface{}) error {
					recovered = true
					return originalFn(r)
				}
			}

			err := yerrors.WithRecover(tt.fn, recoverFn)
			
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("WithRecover() error mismatch (-want +got):\n%s", diff)
			}
			
			if tt.expectRecover && !recovered {
				t.Error("expected recovery function to be called")
			}
			if !tt.expectRecover && recovered {
				t.Error("recovery function was called unexpectedly")
			}
		})
	}
}

func TestAs(t *testing.T) {
	ysoErr := yerrors.NewError(yerrors.CodeNotFound, "not found")
	standardErr := errors.New("standard error")

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "正常系: YSOErrorからYSOErrorへのキャスト",
			err:  ysoErr,
			want: true,
		},
		{
			name: "正常系: 標準エラーからYSOErrorへのキャスト失敗",
			err:  standardErr,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target *yerrors.YSOError
			got := yerrors.As(tt.err, &target)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("As() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIs(t *testing.T) {
	targetErr := errors.New("target error")
	wrappedErr := yerrors.WrapError(yerrors.CodeIO, "wrapped", targetErr)
	otherErr := errors.New("other error")

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "正常系: ラップされたエラーの検出",
			err:    wrappedErr,
			target: targetErr,
			want:   true,
		},
		{
			name:   "正常系: 異なるエラーの比較",
			err:    wrappedErr,
			target: otherErr,
			want:   false,
		},
		{
			name:   "正常系: 同じエラーの比較",
			err:    targetErr,
			target: targetErr,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := yerrors.Is(tt.err, tt.target)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Is() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestYSOError_StackTrace(t *testing.T) {
	err := yerrors.NewError(yerrors.CodeNotFound, "not found")
	stackTrace := err.StackTrace()
	
	if len(stackTrace) == 0 {
		t.Error("expected non-empty stack trace")
	}
}