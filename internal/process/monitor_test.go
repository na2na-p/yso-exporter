package process_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/na2na/yso-exporter/internal/process"
	"github.com/na2na/yso-exporter/internal/process/mocks"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
)

func TestProcessMonitorImpl_Start(t *testing.T) {
	type fields struct {
		eventHandler func(ctrl *gomock.Controller) process.ProcessEventHandler
	}
	type args struct {
		processName string
		ctx         context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "正常系: プロセス監視が正常に開始される",
			fields: fields{
				eventHandler: func(ctrl *gomock.Controller) process.ProcessEventHandler {
					m := mocks.NewMockProcessEventHandler(ctrl)
					return m
				},
			},
			args: args{
				processName: "YsOrigin.exe",
				ctx:         context.Background(),
			},
			wantErr: nil,
		},
		{
			name: "異常系: プロセス名が空の場合エラーが返る",
			fields: fields{
				eventHandler: func(ctrl *gomock.Controller) process.ProcessEventHandler {
					m := mocks.NewMockProcessEventHandler(ctrl)
					return m
				},
			},
			args: args{
				processName: "",
				ctx:         context.Background(),
			},
			wantErr: errors.New("process name cannot be empty"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			
			// 準備: モックをセットアップし、テスト対象インスタンスを生成
			config := &process.MonitorConfig{
				MonitorInterval: 100 * time.Millisecond, // テスト用に短い間隔
				EventHandler:    tt.fields.eventHandler(ctrl),
			}
			finder := process.NewWindowsProcessFinder()
			m := process.NewProcessMonitor(config, finder)
			
			// 実行: テスト対象メソッドを呼び出し
			err := m.Start(tt.args.ctx, tt.args.processName)
			
			// 検証: エラーを比較
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Start() error = nil, wantErr %v", tt.wantErr)
				}
				if diff := cmp.Diff(tt.wantErr.Error(), err.Error()); diff != "" {
					t.Errorf("Start() error mismatch (-want +got):\n%s", diff)
				}
			} else {
				if err != nil {
					t.Fatalf("Start() unexpected error: %v", err)
				}
			}
			
			// クリーンアップ
			if err == nil {
				_ = m.Stop()
			}
		})
	}
}

func TestProcessMonitorImpl_Stop(t *testing.T) {
	type fields struct {
		eventHandler func(ctrl *gomock.Controller) process.ProcessEventHandler
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "正常系: 監視が正常に停止される",
			fields: fields{
				eventHandler: func(ctrl *gomock.Controller) process.ProcessEventHandler {
					m := mocks.NewMockProcessEventHandler(ctrl)
					return m
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			
			// 準備: モックをセットアップし、テスト対象インスタンスを生成
			config := &process.MonitorConfig{
				MonitorInterval: 100 * time.Millisecond,
				EventHandler:    tt.fields.eventHandler(ctrl),
			}
			finder := process.NewWindowsProcessFinder()
			m := process.NewProcessMonitor(config, finder)
			
			// 実行: テスト対象メソッドを呼び出し
			err := m.Stop()
			
			// 検証: エラーを比較
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("Stop() error = nil, wantErr %v", tt.wantErr)
				}
				if diff := cmp.Diff(tt.wantErr.Error(), err.Error()); diff != "" {
					t.Errorf("Stop() error mismatch (-want +got):\n%s", diff)
				}
			} else {
				if err != nil {
					t.Fatalf("Stop() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestProcessMonitorImpl_IsProcessAlive(t *testing.T) {
	type fields struct {
		eventHandler func(ctrl *gomock.Controller) process.ProcessEventHandler
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "正常系: プロセスが生きていない状態でfalseが返る",
			fields: fields{
				eventHandler: func(ctrl *gomock.Controller) process.ProcessEventHandler {
					m := mocks.NewMockProcessEventHandler(ctrl)
					return m
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			
			// 準備: モックをセットアップし、テスト対象インスタンスを生成
			config := &process.MonitorConfig{
				MonitorInterval: 100 * time.Millisecond,
				EventHandler:    tt.fields.eventHandler(ctrl),
			}
			finder := process.NewWindowsProcessFinder()
			m := process.NewProcessMonitor(config, finder)
			
			// 実行: テスト対象メソッドを呼び出し
			got := m.IsProcessAlive()
			
			// 検証: 戻り値を比較
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("IsProcessAlive() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}