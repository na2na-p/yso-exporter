package collector_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/na2na/yso-exporter/internal/collector"
	"github.com/na2na/yso-exporter/internal/collector/mocks"
)

func TestNewSnapshotCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	inner := mocks.NewMockCollector(ctrl)

	sc := collector.NewSnapshotCollector(inner)
	if sc == nil {
		t.Fatal("NewSnapshotCollector() returned nil")
	}

	if got := sc.GetMetrics(); len(got) != 0 {
		t.Errorf("GetMetrics() before first Collect() = %v, want empty", got)
	}
}

func TestSnapshotCollector_Collect(t *testing.T) {
	type fields struct {
		inner func(ctrl *gomock.Controller) collector.Collector
	}
	tests := []struct {
		name        string
		fields      fields
		wantErr     bool
		wantMetrics []collector.Metric
	}{
		{
			name: "正常系: 内側の収集が成功しGetMetricsにスナップショットが反映される",
			fields: fields{
				inner: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().Collect().Return(nil)
					m.EXPECT().GetMetrics().Return([]collector.Metric{
						{Name: "yso_hp", Value: 100, Type: collector.Gauge, Labels: map[string]string{"character": "adol"}},
					})
					return m
				},
			},
			wantErr: false,
			wantMetrics: []collector.Metric{
				{Name: "yso_hp", Value: 100, Type: collector.Gauge, Labels: map[string]string{"character": "adol"}},
			},
		},
		{
			name: "異常系: 内側の収集がエラーの場合エラーを返しスナップショットは空になる",
			fields: fields{
				inner: func(ctrl *gomock.Controller) collector.Collector {
					m := mocks.NewMockCollector(ctrl)
					m.EXPECT().Collect().Return(errors.New("read failed"))
					return m
				},
			},
			wantErr:     true,
			wantMetrics: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			sc := collector.NewSnapshotCollector(tt.fields.inner(ctrl))

			err := sc.Collect()

			if (err != nil) != tt.wantErr {
				t.Fatalf("Collect() error = %v, wantErr %v", err, tt.wantErr)
			}

			got := sc.GetMetrics()
			if diff := cmp.Diff(tt.wantMetrics, got); diff != "" {
				t.Errorf("GetMetrics() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSnapshotCollector_Collect_ClearsSnapshotOnError(t *testing.T) {
	ctrl := gomock.NewController(t)
	inner := mocks.NewMockCollector(ctrl)

	gomock.InOrder(
		inner.EXPECT().Collect().Return(nil),
		inner.EXPECT().GetMetrics().Return([]collector.Metric{
			{Name: "yso_hp", Value: 100, Type: collector.Gauge, Labels: map[string]string{"character": "adol"}},
		}),
		inner.EXPECT().Collect().Return(errors.New("read failed")),
	)

	sc := collector.NewSnapshotCollector(inner)

	if err := sc.Collect(); err != nil {
		t.Fatalf("first Collect() unexpected error: %v", err)
	}

	if err := sc.Collect(); err == nil {
		t.Fatal("second Collect() should return error")
	}

	got := sc.GetMetrics()
	if len(got) != 0 {
		t.Errorf("GetMetrics() after failed Collect() = %v, want empty (stale snapshot must not be served)", got)
	}
}

// TestSnapshotCollector_Collect_SerializesInnerCalls は、複数ゴルーチンから同時にCollect()が
// 呼ばれても、内側のCollect()実行が重複しないことを検証します。
func TestSnapshotCollector_Collect_SerializesInnerCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	inner := mocks.NewMockCollector(ctrl)

	const callCount = 20

	var concurrent int32
	var maxConcurrent int32

	inner.EXPECT().Collect().DoAndReturn(func() error {
		cur := atomic.AddInt32(&concurrent, 1)
		defer atomic.AddInt32(&concurrent, -1)
		for {
			m := atomic.LoadInt32(&maxConcurrent)
			if cur <= m {
				break
			}
			if atomic.CompareAndSwapInt32(&maxConcurrent, m, cur) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		return nil
	}).Times(callCount)
	inner.EXPECT().GetMetrics().Return([]collector.Metric{
		{Name: "yso_hp", Value: 100, Type: collector.Gauge},
	}).Times(callCount)

	sc := collector.NewSnapshotCollector(inner)

	var wg sync.WaitGroup
	for i := 0; i < callCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sc.Collect()
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&maxConcurrent); got > 1 {
		t.Errorf("inner.Collect() ran concurrently: maxConcurrent = %d, want <= 1", got)
	}
}

// TestSnapshotCollector_GetMetrics_DoesNotBlockDuringInnerCollect は、Collect()内側の
// メモリ読み取りが実行中でも、GetMetrics()(読み手)がブロックされないことを検証します。
// スナップショットの読み書きロックとCollect直列化用の排他が分離されていることの確認です。
func TestSnapshotCollector_GetMetrics_DoesNotBlockDuringInnerCollect(t *testing.T) {
	ctrl := gomock.NewController(t)
	inner := mocks.NewMockCollector(ctrl)

	started := make(chan struct{})
	release := make(chan struct{})

	inner.EXPECT().Collect().DoAndReturn(func() error {
		close(started)
		<-release
		return nil
	})
	inner.EXPECT().GetMetrics().Return([]collector.Metric{
		{Name: "yso_hp", Value: 100, Type: collector.Gauge},
	})

	sc := collector.NewSnapshotCollector(inner)

	collectDone := make(chan struct{})
	go func() {
		_ = sc.Collect()
		close(collectDone)
	}()

	<-started

	getMetricsDone := make(chan struct{})
	go func() {
		sc.GetMetrics()
		close(getMetricsDone)
	}()

	select {
	case <-getMetricsDone:
		// 進行中のCollect()の完了を待たずにGetMetrics()が返っていればOK。
	case <-time.After(1 * time.Second):
		t.Fatal("GetMetrics() blocked while inner.Collect() was in progress")
	}

	close(release)
	<-collectDone
}

func TestSnapshotCollector_GetMetrics_ReturnsIndependentCopy(t *testing.T) {
	ctrl := gomock.NewController(t)
	inner := mocks.NewMockCollector(ctrl)
	inner.EXPECT().Collect().Return(nil)
	inner.EXPECT().GetMetrics().Return([]collector.Metric{
		{Name: "yso_hp", Value: 100, Type: collector.Gauge, Labels: map[string]string{"character": "adol"}},
	})

	sc := collector.NewSnapshotCollector(inner)
	if err := sc.Collect(); err != nil {
		t.Fatalf("Collect() unexpected error: %v", err)
	}

	got := sc.GetMetrics()
	got[0].Value = 9999
	got[0].Labels["character"] = "mutated"

	got2 := sc.GetMetrics()
	if diff := cmp.Diff(100.0, got2[0].Value); diff != "" {
		t.Errorf("internal snapshot value mutated by caller: %v", diff)
	}
	if diff := cmp.Diff("adol", got2[0].Labels["character"]); diff != "" {
		t.Errorf("internal snapshot label mutated by caller: %v", diff)
	}
}

func TestSnapshotCollector_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	inner := mocks.NewMockCollector(ctrl)
	inner.EXPECT().Collect().Return(nil).AnyTimes()
	inner.EXPECT().GetMetrics().Return([]collector.Metric{
		{Name: "yso_hp", Value: 100, Type: collector.Gauge},
	}).AnyTimes()

	sc := collector.NewSnapshotCollector(inner)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = sc.Collect()
		}()
		go func() {
			defer wg.Done()
			_ = sc.GetMetrics()
		}()
	}
	wg.Wait()
}
