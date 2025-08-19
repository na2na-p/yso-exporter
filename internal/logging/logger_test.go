package logging_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/na2na/yso-exporter/internal/logging"
)

func TestNewZapLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     logging.LogLevel
		wantLevel logging.LogLevel
	}{
		{
			name:      "正常系: DEBUG レベルで初期化される",
			level:     logging.LevelDebug,
			wantLevel: logging.LevelDebug,
		},
		{
			name:      "正常系: INFO レベルで初期化される",
			level:     logging.LevelInfo,
			wantLevel: logging.LevelInfo,
		},
		{
			name:      "正常系: WARN レベルで初期化される",
			level:     logging.LevelWarn,
			wantLevel: logging.LevelWarn,
		},
		{
			name:      "正常系: ERROR レベルで初期化される",
			level:     logging.LevelError,
			wantLevel: logging.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logging.NewZapLogger(tt.level, &buf)

			if logger == nil {
				t.Fatal("NewZapLogger() returned nil")
			}

			// ログレベルが正しく設定されていることを確認するため、
			// 各レベルでログを出力して期待するレベルのみが出力されることを検証
			testMessage := "test message"
			
			logger.Debug(testMessage)
			logger.Info(testMessage)
			logger.Warn(testMessage)
			logger.Error(testMessage)

			output := buf.String()
			lines := strings.Split(strings.TrimSpace(output), "\n")
			
			// 空行を除去
			var validLines []string
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					validLines = append(validLines, line)
				}
			}

			// ログレベルに応じて期待する行数を計算
			var expectedLineCount int
			switch tt.level {
			case logging.LevelDebug:
				expectedLineCount = 4 // DEBUG, INFO, WARN, ERROR
			case logging.LevelInfo:
				expectedLineCount = 3 // INFO, WARN, ERROR
			case logging.LevelWarn:
				expectedLineCount = 2 // WARN, ERROR
			case logging.LevelError:
				expectedLineCount = 1 // ERROR
			}

			if len(validLines) != expectedLineCount {
				t.Errorf("expected %d log lines, got %d. Output: %s", expectedLineCount, len(validLines), output)
			}
		})
	}
}

func TestZapLogger_Debug(t *testing.T) {
	tests := []struct {
		name        string
		level       logging.LogLevel
		msg         string
		wantLogged  bool
		wantMessage string
	}{
		{
			name:        "正常系: DEBUG レベルで DEBUG メッセージがログ出力される",
			level:       logging.LevelDebug,
			msg:         "debug message",
			wantLogged:  true,
			wantMessage: "debug message",
		},
		{
			name:       "正常系: INFO レベルで DEBUG メッセージが出力されない",
			level:      logging.LevelInfo,
			msg:        "debug message",
			wantLogged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logging.NewZapLogger(tt.level, &buf)

			logger.Debug(tt.msg)

			output := buf.String()
			if tt.wantLogged {
				if output == "" {
					t.Error("expected log output, but got none")
				}
				
				var logEntry map[string]interface{}
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
					if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
						t.Fatalf("failed to parse JSON log: %v", err)
					}
					
					if msg, ok := logEntry["msg"]; ok {
						if diff := cmp.Diff(tt.wantMessage, msg); diff != "" {
							t.Errorf("message mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Error("log entry missing 'msg' field")
					}
				}
			} else {
				if output != "" {
					t.Errorf("expected no log output, but got: %s", output)
				}
			}
		})
	}
}

func TestZapLogger_Info(t *testing.T) {
	tests := []struct {
		name        string
		level       logging.LogLevel
		msg         string
		wantLogged  bool
		wantMessage string
	}{
		{
			name:        "正常系: DEBUG レベルで INFO メッセージがログ出力される",
			level:       logging.LevelDebug,
			msg:         "info message",
			wantLogged:  true,
			wantMessage: "info message",
		},
		{
			name:        "正常系: INFO レベルで INFO メッセージがログ出力される",
			level:       logging.LevelInfo,
			msg:         "info message",
			wantLogged:  true,
			wantMessage: "info message",
		},
		{
			name:       "正常系: WARN レベルで INFO メッセージが出力されない",
			level:      logging.LevelWarn,
			msg:        "info message",
			wantLogged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logging.NewZapLogger(tt.level, &buf)

			logger.Info(tt.msg)

			output := buf.String()
			if tt.wantLogged {
				if output == "" {
					t.Error("expected log output, but got none")
				}
				
				var logEntry map[string]interface{}
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
					if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
						t.Fatalf("failed to parse JSON log: %v", err)
					}
					
					if msg, ok := logEntry["msg"]; ok {
						if diff := cmp.Diff(tt.wantMessage, msg); diff != "" {
							t.Errorf("message mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Error("log entry missing 'msg' field")
					}
				}
			} else {
				if output != "" {
					t.Errorf("expected no log output, but got: %s", output)
				}
			}
		})
	}
}

func TestZapLogger_Infof(t *testing.T) {
	tests := []struct {
		name        string
		level       logging.LogLevel
		format      string
		args        []interface{}
		wantLogged  bool
		wantMessage string
	}{
		{
			name:        "正常系: フォーマット済みメッセージがログ出力される",
			level:       logging.LevelInfo,
			format:      "Process %s connected with PID %d",
			args:        []interface{}{"YsOrigin.exe", 1234},
			wantLogged:  true,
			wantMessage: "Process YsOrigin.exe connected with PID 1234",
		},
		{
			name:       "正常系: WARN レベルで INFO メッセージが出力されない",
			level:      logging.LevelWarn,
			format:     "Process %s connected",
			args:       []interface{}{"YsOrigin.exe"},
			wantLogged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logging.NewZapLogger(tt.level, &buf)

			logger.Infof(tt.format, tt.args...)

			output := buf.String()
			if tt.wantLogged {
				if output == "" {
					t.Error("expected log output, but got none")
				}
				
				var logEntry map[string]interface{}
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
					if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
						t.Fatalf("failed to parse JSON log: %v", err)
					}
					
					if msg, ok := logEntry["msg"]; ok {
						if diff := cmp.Diff(tt.wantMessage, msg); diff != "" {
							t.Errorf("message mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Error("log entry missing 'msg' field")
					}
				}
			} else {
				if output != "" {
					t.Errorf("expected no log output, but got: %s", output)
				}
			}
		})
	}
}

func TestZapLogger_InfoWith(t *testing.T) {
	tests := []struct {
		name        string
		level       logging.LogLevel
		msg         string
		fields      map[string]interface{}
		wantLogged  bool
		wantMessage string
		wantFields  map[string]interface{}
	}{
		{
			name:   "正常系: フィールド付きメッセージがログ出力される",
			level:  logging.LevelInfo,
			msg:    "Memory read operation",
			fields: map[string]interface{}{"address": "0x12345678", "size": 4, "value": 100},
			wantLogged: true,
			wantMessage: "Memory read operation",
			wantFields: map[string]interface{}{"address": "0x12345678", "size": float64(4), "value": float64(100)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logging.NewZapLogger(tt.level, &buf)

			logger.InfoWith(tt.msg, tt.fields)

			output := buf.String()
			if tt.wantLogged {
				if output == "" {
					t.Error("expected log output, but got none")
				}
				
				var logEntry map[string]interface{}
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
					if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
						t.Fatalf("failed to parse JSON log: %v", err)
					}
					
					// メッセージの検証
					if msg, ok := logEntry["msg"]; ok {
						if diff := cmp.Diff(tt.wantMessage, msg); diff != "" {
							t.Errorf("message mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Error("log entry missing 'msg' field")
					}
					
					// フィールドの検証
					for key, expectedValue := range tt.wantFields {
						if actualValue, ok := logEntry[key]; ok {
							if diff := cmp.Diff(expectedValue, actualValue); diff != "" {
								t.Errorf("field %s mismatch (-want +got):\n%s", key, diff)
							}
						} else {
							t.Errorf("log entry missing field '%s'", key)
						}
					}
				}
			}
		})
	}
}

func TestZapLogger_WithFields(t *testing.T) {
	tests := []struct {
		name           string
		level          logging.LogLevel
		baseFields     map[string]interface{}
		msg            string
		wantLogged     bool
		wantMessage    string
		wantFields     map[string]interface{}
	}{
		{
			name:       "正常系: WithFields で設定されたフィールドがログに含まれる",
			level:      logging.LevelInfo,
			baseFields: map[string]interface{}{"component": "memory_reader", "pid": 1234},
			msg:        "Operation completed",
			wantLogged: true,
			wantMessage: "Operation completed",
			wantFields: map[string]interface{}{"component": "memory_reader", "pid": float64(1234)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logging.NewZapLogger(tt.level, &buf)

			fieldLogger := logger.WithFields(tt.baseFields)
			fieldLogger.Info(tt.msg)

			output := buf.String()
			if tt.wantLogged {
				if output == "" {
					t.Error("expected log output, but got none")
				}
				
				var logEntry map[string]interface{}
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
					if err := json.Unmarshal([]byte(lines[0]), &logEntry); err != nil {
						t.Fatalf("failed to parse JSON log: %v", err)
					}
					
					// メッセージの検証
					if msg, ok := logEntry["msg"]; ok {
						if diff := cmp.Diff(tt.wantMessage, msg); diff != "" {
							t.Errorf("message mismatch (-want +got):\n%s", diff)
						}
					} else {
						t.Error("log entry missing 'msg' field")
					}
					
					// フィールドの検証
					for key, expectedValue := range tt.wantFields {
						if actualValue, ok := logEntry[key]; ok {
							if diff := cmp.Diff(expectedValue, actualValue); diff != "" {
								t.Errorf("field %s mismatch (-want +got):\n%s", key, diff)
							}
						} else {
							t.Errorf("log entry missing field '%s'", key)
						}
					}
				}
			}
		})
	}
}

func TestZapLogger_SetLevel(t *testing.T) {
	tests := []struct {
		name           string
		initialLevel   logging.LogLevel
		newLevel       logging.LogLevel
		testMessage    string
		expectLogged   bool
	}{
		{
			name:         "正常系: DEBUG から INFO に変更後、DEBUG メッセージが出力されない",
			initialLevel: logging.LevelDebug,
			newLevel:     logging.LevelInfo,
			testMessage:  "debug message",
			expectLogged: false,
		},
		{
			name:         "正常系: WARN から DEBUG に変更後、DEBUG メッセージが出力される",
			initialLevel: logging.LevelWarn,
			newLevel:     logging.LevelDebug,
			testMessage:  "debug message",
			expectLogged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logging.NewZapLogger(tt.initialLevel, &buf)

			// レベルを変更
			logger.SetLevel(tt.newLevel)

			// バッファをクリア
			buf.Reset()

			// テストメッセージを出力
			logger.Debug(tt.testMessage)

			output := buf.String()
			if tt.expectLogged {
				if output == "" {
					t.Error("expected log output after level change, but got none")
				}
			} else {
				if output != "" {
					t.Errorf("expected no log output after level change, but got: %s", output)
				}
			}
		})
	}
}