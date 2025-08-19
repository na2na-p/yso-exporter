package collector_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/na2na/yso-exporter/internal/collector"
	"github.com/na2na/yso-exporter/internal/collector/mocks"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
)

func TestNewPoller(t *testing.T) {
	type args struct {
		collector collector.Collector
	}
	tests := []struct {
		name string
		args args
		want *collector.PollerImpl
	}{
		{
			name: "正常系: 有効なCollectorでPollerが作成される",
			args: args{
				collector: &mockCollector{},
			},
			want: &collector.PollerImpl{}, // 具体的な値の検証は別途行う
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collector.NewPoller(tt.args.collector)
			if got == nil {
				t.Fatal("NewPoller() returned nil")
			}
			if got.IsRunning() {
				t.Error("newly created poller should not be running")
			}
		})
	}
}

func TestPollerImpl_Start(t *testing.T) {
	type fields struct {
		collector func(ctrl *gomock.Controller) collector.Collector
	}
	type args struct {
		interval time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "正常系: 有効な間隔でポーリングが開始される",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					// ポーリング中にCollectが呼ばれることを期待
					m.EXPECT().Collect().Return(nil).AnyTimes()
					return m
				},
			},
			args: args{
				interval: 10 * time.Millisecond,
			},
			wantErr: nil,
		},
		{
			name: "異常系: 無効な間隔（0）でエラーが返る",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					// エラーケースではCollectは呼ばれない
					return m
				},
			},
			args: args{
				interval: 0,
			},
			wantErr: errors.New("interval must be positive: 0s"),
		},
		{
			name: "異常系: 負の間隔でエラーが返る",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					return m
				},
			},
			args: args{
				interval: -1 * time.Millisecond,
			},
			wantErr: errors.New("interval must be positive: -1ms"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			
			p := collector.NewPoller(tt.fields.collector(ctrl))
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			
			err := p.Start(ctx, tt.args.interval)
			
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
				
				// ポーリングが開始されたことを確認
				if !p.IsRunning() {
					t.Error("poller should be running after Start()")
				}
				
				// 少し待ってからStopを呼ぶ
				time.Sleep(30 * time.Millisecond)
				if stopErr := p.Stop(); stopErr != nil {
					t.Errorf("Stop() error = %v", stopErr)
				}
			}
		})
	}
}

func TestPollerImpl_Start_AlreadyRunning(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mocks.NewMockCollector(ctrl)
	m.EXPECT().Collect().Return(nil).AnyTimes()
	
	p := collector.NewPoller(m)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// 最初のStart
	err := p.Start(ctx, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("first Start() failed: %v", err)
	}
	
	// 2回目のStart（エラーになるはず）
	err = p.Start(ctx, 10*time.Millisecond)
	if err == nil {
		t.Fatal("second Start() should return error")
	}
	
	expectedErr := "poller is already running"
	if diff := cmp.Diff(expectedErr, err.Error()); diff != "" {
		t.Errorf("Start() error mismatch (-want +got):\n%s", diff)
	}
	
	// クリーンアップ
	if stopErr := p.Stop(); stopErr != nil {
		t.Errorf("Stop() error = %v", stopErr)
	}
}

func TestPollerImpl_Stop(t *testing.T) {
	type fields struct {
		collector func(ctrl *gomock.Controller) collector.Collector
	}
	tests := []struct {
		name      string
		fields    fields
		setupFunc func(p collector.MetricPoller) error // ポーラーの初期状態を設定
		wantErr   error
	}{
		{
			name: "正常系: 実行中のポーリングが停止される",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().Collect().Return(nil).AnyTimes()
					return m
				},
			},
			setupFunc: func(p collector.MetricPoller) error {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				return p.Start(ctx, 10*time.Millisecond)
			},
			wantErr: nil,
		},
		{
			name: "異常系: 実行されていないポーリングの停止でエラーが返る",
			fields: fields{
				collector: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					return m
				},
			},
			setupFunc: func(p collector.MetricPoller) error {
				// 何もしない（ポーラーは開始されていない）
				return nil
			},
			wantErr: errors.New("poller is not running"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			
			p := collector.NewPoller(tt.fields.collector(ctrl))
			
			// セットアップ
			if setupErr := tt.setupFunc(p); setupErr != nil {
				t.Fatalf("setup failed: %v", setupErr)
			}
			
			err := p.Stop()
			
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
				
				// ポーリングが停止されたことを確認
				if p.IsRunning() {
					t.Error("poller should not be running after Stop()")
				}
			}
		})
	}
}

func TestPollerImpl_IsRunning(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mocks.NewMockCollector(ctrl)
	m.EXPECT().Collect().Return(nil).AnyTimes()
	
	p := collector.NewPoller(m)
	
	// 初期状態：実行されていない
	if p.IsRunning() {
		t.Error("newly created poller should not be running")
	}
	
	// Start後：実行中
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	if err := p.Start(ctx, 10*time.Millisecond); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	
	if !p.IsRunning() {
		t.Error("poller should be running after Start()")
	}
	
	// Stop後：実行されていない
	if err := p.Stop(); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}
	
	if p.IsRunning() {
		t.Error("poller should not be running after Stop()")
	}
}

func TestPollerImpl_PollingWithCollectorError(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mocks.NewMockCollector(ctrl)
	
	// Collectでエラーが発生するが、ポーリングは継続することを確認
	m.EXPECT().Collect().Return(errors.New("collection failed")).AnyTimes()
	
	p := collector.NewPoller(m)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	err := p.Start(ctx, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	
	// しばらく待ってポーリングが継続されることを確認
	time.Sleep(30 * time.Millisecond)
	
	if !p.IsRunning() {
		t.Error("poller should continue running even when Collect() returns error")
	}
	
	// contextのキャンセルでポーリングループは停止するが、
	// IsRunning()はStop()が呼ばれるまでtrueのまま
	cancel()
	
	// 少し待ってからStop（明示的に呼ぶ必要がある）
	time.Sleep(20 * time.Millisecond)
	
	if !p.IsRunning() {
		t.Error("poller should still report as running until Stop() is called")
	}
	
	// 明示的にStopを呼ぶ
	if stopErr := p.Stop(); stopErr != nil {
		t.Errorf("Stop() failed: %v", stopErr)
	}
	
	if p.IsRunning() {
		t.Error("poller should not be running after Stop()")
	}
}

func TestPollerImpl_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mocks.NewMockCollector(ctrl)
	m.EXPECT().Collect().Return(nil).AnyTimes()
	
	p := collector.NewPoller(m)
	ctx, cancel := context.WithCancel(context.Background())
	
	err := p.Start(ctx, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	
	if !p.IsRunning() {
		t.Error("poller should be running after Start()")
	}
	
	// contextをキャンセル
	cancel()
	
	// ポーリングループは内部的に停止するが、
	// IsRunning()はStop()が呼ばれるまでtrueのまま
	time.Sleep(50 * time.Millisecond)
	
	if !p.IsRunning() {
		t.Error("poller should still report as running until Stop() is called")
	}
	
	// Stop()を呼ぶと即座に完了する（ループは既に停止している）
	if stopErr := p.Stop(); stopErr != nil {
		t.Errorf("Stop() failed: %v", stopErr)
	}
	
	if p.IsRunning() {
		t.Error("poller should not be running after Stop()")
	}
}

// mockCollector は、テスト用の簡単なCollector実装です。
type mockCollector struct{}

func (m *mockCollector) Collect() error {
	return nil
}

func (m *mockCollector) GetMetrics() []collector.Metric {
	return []collector.Metric{}
}