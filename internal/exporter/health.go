package exporter

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/na2na/yso-exporter/internal/process"
)

// HealthHandler は、ヘルスチェックエンドポイントを提供するHTTPハンドラーです。
// カプセル化の原則に従い、ヘルスチェック機能という単一の目的を持ちます。
type HealthHandler struct {
	processMonitor process.ProcessMonitor
}

// HealthResponse は、ヘルスチェックのレスポンス形式を定義します。
type HealthResponse struct {
	Status           string `json:"status"`
	ProcessConnected bool   `json:"process_connected"`
}

// NewHealthHandler は、HealthHandlerの新しいインスタンスを作成します。
// 関心の分離の原則に従い、ProcessMonitorとの依存関係をコンストラクタで注入します。
func NewHealthHandler(processMonitor process.ProcessMonitor) *HealthHandler {
	return &HealthHandler{
		processMonitor: processMonitor,
	}
}

// ServeHTTP は、HTTPリクエストを処理してヘルスチェック結果をJSON形式で返します。
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// HTTPメソッドの検証
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, "Method not allowed")
		return
	}

	// プロセス状態を確認
	isProcessAlive := h.processMonitor.IsProcessAlive()

	// レスポンスデータを準備
	var response HealthResponse
	var statusCode int

	if isProcessAlive {
		// プロセス接続時: 正常なレスポンス
		response = HealthResponse{
			Status:           "ok",
			ProcessConnected: true,
		}
		statusCode = http.StatusOK
	} else {
		// プロセス未接続時: エラーレスポンス
		response = HealthResponse{
			Status:           "error",
			ProcessConnected: false,
		}
		statusCode = http.StatusServiceUnavailable
	}

	// Content-Typeヘッダーを設定
	w.Header().Set("Content-Type", "application/json")

	// ステータスコードを設定
	w.WriteHeader(statusCode)

	// JSONレスポンスを出力
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// JSONエンコードエラーの場合、プレーンテキストでエラーを返す
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
	}
}