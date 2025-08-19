package exporter_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/na2na/yso-exporter/internal/exporter"
	"github.com/na2na/yso-exporter/internal/process/mocks"
	"go.uber.org/mock/gomock"
)

func TestHealthHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		processMonitor func(ctrl *gomock.Controller) *mocks.MockProcessMonitor
	}
	type args struct {
		method string
		path   string
	}
	type want struct {
		statusCode   int
		contentType  string
		responseBody map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "正常系: プロセス接続時、status=ok、process_connected=trueが返る",
			fields: fields{
				processMonitor: func(ctrl *gomock.Controller) *mocks.MockProcessMonitor {
					m := mocks.NewMockProcessMonitor(ctrl)
					m.EXPECT().IsProcessAlive().Return(true)
					return m
				},
			},
			args: args{
				method: http.MethodGet,
				path:   "/health",
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				responseBody: map[string]interface{}{
					"status":            "ok",
					"process_connected": true,
				},
			},
		},
		{
			name: "正常系: プロセス未接続時、status=error、process_connected=falseが返る",
			fields: fields{
				processMonitor: func(ctrl *gomock.Controller) *mocks.MockProcessMonitor {
					m := mocks.NewMockProcessMonitor(ctrl)
					m.EXPECT().IsProcessAlive().Return(false)
					return m
				},
			},
			args: args{
				method: http.MethodGet,
				path:   "/health",
			},
			want: want{
				statusCode:  http.StatusServiceUnavailable,
				contentType: "application/json",
				responseBody: map[string]interface{}{
					"status":            "error",
					"process_connected": false,
				},
			},
		},
		{
			name: "異常系: GET以外のメソッドの場合、405エラーが返る",
			fields: fields{
				processMonitor: func(ctrl *gomock.Controller) *mocks.MockProcessMonitor {
					m := mocks.NewMockProcessMonitor(ctrl)
					// POSTの場合はIsProcessAlive()は呼ばれない
					return m
				},
			},
			args: args{
				method: http.MethodPost,
				path:   "/health",
			},
			want: want{
				statusCode:  http.StatusMethodNotAllowed,
				contentType: "text/plain; charset=utf-8",
				responseBody: map[string]interface{}{
					"error": "Method not allowed",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			// 準備: モックをセットアップし、テスト対象インスタンスを生成
			processMonitor := tt.fields.processMonitor(ctrl)
			h := exporter.NewHealthHandler(processMonitor)

			// HTTPリクエストとレスポンスレコーダーを準備
			req := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			w := httptest.NewRecorder()

			// 実行: テスト対象メソッドを呼び出し
			h.ServeHTTP(w, req)

			// 検証: ステータスコードを確認
			if diff := cmp.Diff(tt.want.statusCode, w.Code); diff != "" {
				t.Errorf("HealthHandler.ServeHTTP() status code mismatch (-want +got):\n%s", diff)
			}

			// 検証: Content-Typeヘッダーを確認
			if diff := cmp.Diff(tt.want.contentType, w.Header().Get("Content-Type")); diff != "" {
				t.Errorf("HealthHandler.ServeHTTP() Content-Type mismatch (-want +got):\n%s", diff)
			}

			// 検証: レスポンスボディを確認
			if tt.want.contentType == "application/json" {
				var got map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("HealthHandler.ServeHTTP() failed to unmarshal response body: %v", err)
				}
				if diff := cmp.Diff(tt.want.responseBody, got); diff != "" {
					t.Errorf("HealthHandler.ServeHTTP() response body mismatch (-want +got):\n%s", diff)
				}
			} else {
				// テキストレスポンスの場合
				got := w.Body.String()
				want := tt.want.responseBody["error"].(string)
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("HealthHandler.ServeHTTP() response body mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestNewHealthHandler(t *testing.T) {
	type args struct {
		processMonitor func(ctrl *gomock.Controller) *mocks.MockProcessMonitor
	}
	tests := []struct {
		name string
		args args
		want bool // HealthHandlerが正常に作成されるかのフラグ
	}{
		{
			name: "正常系: ProcessMonitorを指定してHealthHandlerが正常に作成される",
			args: args{
				processMonitor: func(ctrl *gomock.Controller) *mocks.MockProcessMonitor {
					return mocks.NewMockProcessMonitor(ctrl)
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			// 準備: モックをセットアップ
			processMonitor := tt.args.processMonitor(ctrl)

			// 実行: テスト対象メソッドを呼び出し
			got := exporter.NewHealthHandler(processMonitor)

			// 検証: HealthHandlerが正常に作成されるか確認
			if (got != nil) != tt.want {
				t.Errorf("NewHealthHandler() = %v, want %v", got != nil, tt.want)
			}
		})
	}
}