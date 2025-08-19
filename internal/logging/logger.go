package logging

import (
	"fmt"
	"io"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate mockgen -source=logger.go -destination=mocks/mock_logger.go -package=mocks

// LogLevel は ログレベルを表します。
type LogLevel string

const (
	// LevelDebug は DEBUG レベルです。
	LevelDebug LogLevel = "DEBUG"
	// LevelInfo は INFO レベルです。
	LevelInfo LogLevel = "INFO"
	// LevelWarn は WARN レベルです。
	LevelWarn LogLevel = "WARN"
	// LevelError は ERROR レベルです。
	LevelError LogLevel = "ERROR"
)

// Logger は構造化ログ出力のためのインターフェースです。
// 目的: デバッグとトラブルシューティングのためのログ出力機能を提供する
type Logger interface {
	// Debug は DEBUG レベルでメッセージをログ出力します。
	Debug(msg string)
	
	// Debugf は DEBUG レベルでフォーマット済みメッセージをログ出力します。
	Debugf(format string, args ...interface{})
	
	// DebugWith は DEBUG レベルでフィールド付きメッセージをログ出力します。
	DebugWith(msg string, fields map[string]interface{})
	
	// Info は INFO レベルでメッセージをログ出力します。
	Info(msg string)
	
	// Infof は INFO レベルでフォーマット済みメッセージをログ出力します。
	Infof(format string, args ...interface{})
	
	// InfoWith は INFO レベルでフィールド付きメッセージをログ出力します。
	InfoWith(msg string, fields map[string]interface{})
	
	// Warn は WARN レベルでメッセージをログ出力します。
	Warn(msg string)
	
	// Warnf は WARN レベルでフォーマット済みメッセージをログ出力します。
	Warnf(format string, args ...interface{})
	
	// WarnWith は WARN レベルでフィールド付きメッセージをログ出力します。
	WarnWith(msg string, fields map[string]interface{})
	
	// Error は ERROR レベルでメッセージをログ出力します。
	Error(msg string)
	
	// Errorf は ERROR レベルでフォーマット済みメッセージをログ出力します。
	Errorf(format string, args ...interface{})
	
	// ErrorWith は ERROR レベルでフィールド付きメッセージをログ出力します。
	ErrorWith(msg string, fields map[string]interface{})
	
	// WithFields は指定されたフィールドを持つ新しい Logger を返します。
	WithFields(fields map[string]interface{}) Logger
	
	// SetLevel は ログレベルを設定します。
	SetLevel(level LogLevel)
}

// zapLogger は zap を使用した Logger の実装です。
// 目的: 構造化ログ（JSON形式）出力とログレベル制御機能を提供する
type zapLogger struct {
	logger *zap.Logger
	level  *zap.AtomicLevel
	mutex  sync.RWMutex
}

// NewZapLogger は指定されたログレベルと出力先で新しい zapLogger を作成します。
func NewZapLogger(level LogLevel, writer io.Writer) Logger {
	atomicLevel := zap.NewAtomicLevel()
	zapLevel := logLevelToZapLevel(level)
	atomicLevel.SetLevel(zapLevel)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(writer),
		atomicLevel,
	)

	logger := zap.New(core)
	return &zapLogger{
		logger: logger,
		level:  &atomicLevel,
	}
}

// logLevelToZapLevel は LogLevel を zapcore.Level に変換します。
func logLevelToZapLevel(level LogLevel) zapcore.Level {
	switch level {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Debug は DEBUG レベルでメッセージをログ出力します。
func (z *zapLogger) Debug(msg string) {
	z.logger.Debug(msg)
}

// Debugf は DEBUG レベルでフォーマット済みメッセージをログ出力します。
func (z *zapLogger) Debugf(format string, args ...interface{}) {
	z.logger.Debug(fmt.Sprintf(format, args...))
}

// DebugWith は DEBUG レベルでフィールド付きメッセージをログ出力します。
func (z *zapLogger) DebugWith(msg string, fields map[string]interface{}) {
	z.logger.Debug(msg, z.mapToZapFields(fields)...)
}

// Info は INFO レベルでメッセージをログ出力します。
func (z *zapLogger) Info(msg string) {
	z.logger.Info(msg)
}

// Infof は INFO レベルでフォーマット済みメッセージをログ出力します。
func (z *zapLogger) Infof(format string, args ...interface{}) {
	z.logger.Info(fmt.Sprintf(format, args...))
}

// InfoWith は INFO レベルでフィールド付きメッセージをログ出力します。
func (z *zapLogger) InfoWith(msg string, fields map[string]interface{}) {
	z.logger.Info(msg, z.mapToZapFields(fields)...)
}

// Warn は WARN レベルでメッセージをログ出力します。
func (z *zapLogger) Warn(msg string) {
	z.logger.Warn(msg)
}

// Warnf は WARN レベルでフォーマット済みメッセージをログ出力します。
func (z *zapLogger) Warnf(format string, args ...interface{}) {
	z.logger.Warn(fmt.Sprintf(format, args...))
}

// WarnWith は WARN レベルでフィールド付きメッセージをログ出力します。
func (z *zapLogger) WarnWith(msg string, fields map[string]interface{}) {
	z.logger.Warn(msg, z.mapToZapFields(fields)...)
}

// Error は ERROR レベルでメッセージをログ出力します。
func (z *zapLogger) Error(msg string) {
	z.logger.Error(msg)
}

// Errorf は ERROR レベルでフォーマット済みメッセージをログ出力します。
func (z *zapLogger) Errorf(format string, args ...interface{}) {
	z.logger.Error(fmt.Sprintf(format, args...))
}

// ErrorWith は ERROR レベルでフィールド付きメッセージをログ出力します。
func (z *zapLogger) ErrorWith(msg string, fields map[string]interface{}) {
	z.logger.Error(msg, z.mapToZapFields(fields)...)
}

// WithFields は指定されたフィールドを持つ新しい Logger を返します。
func (z *zapLogger) WithFields(fields map[string]interface{}) Logger {
	newLogger := z.logger.With(z.mapToZapFields(fields)...)
	return &zapLogger{
		logger: newLogger,
		level:  z.level,
	}
}

// SetLevel は ログレベルを設定します。
func (z *zapLogger) SetLevel(level LogLevel) {
	z.mutex.Lock()
	defer z.mutex.Unlock()
	zapLevel := logLevelToZapLevel(level)
	z.level.SetLevel(zapLevel)
}

// mapToZapFields は map[string]interface{} を zap.Field のスライスに変換します。
func (z *zapLogger) mapToZapFields(fields map[string]interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return zapFields
}