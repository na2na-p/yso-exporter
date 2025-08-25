package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/na2na/yso-exporter/internal/collector"
	"github.com/na2na/yso-exporter/internal/config"
	"github.com/na2na/yso-exporter/internal/debug"
	"github.com/na2na/yso-exporter/internal/exporter"
	"github.com/na2na/yso-exporter/internal/memory"
	"github.com/na2na/yso-exporter/internal/process"
	"github.com/na2na/yso-exporter/internal/version"
)

// Application represents the main application structure
// カプセル化の原則に従い、アプリケーション全体を管理する単一の責任を持つ
type Application struct {
	flags          *config.Flags
	memoryReader   memory.MemoryReader
	collector      collector.Collector
	poller         *collector.PollerImpl
	processMonitor process.ProcessMonitor
	httpServer     *http.Server
	ctx            context.Context
	cancel         context.CancelFunc
}

// ProcessEventHandler implements process.ProcessEventHandler interface
// 関心の分離の原則に従い、プロセスイベント処理を専門化
type ProcessEventHandler struct {
	memoryReader memory.MemoryReader
	processName  string
	callback     ProcessStatusCallback
}

// ProcessStatusCallback is called when process status changes
type ProcessStatusCallback interface {
	OnProcessConnected(processName string)
	OnProcessLost(processName string)
}

func (peh *ProcessEventHandler) OnProcessFound(processName string) error {
	log.Printf("Process found: %s", processName)
	if err := peh.memoryReader.OpenProcess(processName); err != nil {
		log.Printf("Failed to open process %s: %v", processName, err)
		return err
	}
	log.Printf("Successfully connected to process: %s", processName)
	if peh.callback != nil {
		peh.callback.OnProcessConnected(processName)
	}
	return nil
}

func (peh *ProcessEventHandler) OnProcessLost(processName string) error {
	log.Printf("Process lost: %s", processName)
	if err := peh.memoryReader.Close(); err != nil {
		log.Printf("Failed to close memory reader: %v", err)
	}
	if peh.callback != nil {
		peh.callback.OnProcessLost(processName)
	}
	return nil
}

func main() {
	flags := config.ParseFlags()

	// Initialize debug logging
	debug.Init(flags.Debug)
	defer debug.Close()
	
	debug.Log("Starting YSO Exporter with flags: %+v", flags)

	if flags.GUI {
		debug.Log("Launching GUI mode")
		runGUI()
		return
	}

	if flags.ListVersions {
		listAvailableVersions()
		return
	}

	selectedVersion := flags.Version
	if selectedVersion == "" {
		selectedVersion = selectVersionInteractively()
		if selectedVersion == "" {
			return
		}
		flags.Version = selectedVersion
	}

	app, err := NewApplication(flags)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	defer app.Cleanup()

	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	app.WaitForShutdown()
}

// NewApplication creates and initializes a new Application instance
// 関心の分離の原則に従い、各コンポーネントの初期化を管理
func NewApplication(flags *config.Flags) (*Application, error) {
	// Context設定
	ctx, cancel := context.WithCancel(context.Background())

	app := &Application{
		flags:  flags,
		ctx:    ctx,
		cancel: cancel,
	}

	// AddressResolverの作成（組み込み設定から直接作成）
	addressResolver, err := version.NewConfigAddressResolverByVersion(flags.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to create address resolver: %w", err)
	}

	// ProcessFinderの初期化
	processFinder := process.NewWindowsProcessFinder()

	// MemoryReaderの初期化
	memoryReader := memory.NewWindowsMemoryReader(processFinder)
	app.memoryReader = memoryReader

	// Collectorの初期化
	basicCollector := collector.NewBasicCollector(addressResolver, memoryReader)
	app.collector = basicCollector

	// Pollerの初期化
	app.poller = collector.NewPoller(basicCollector)

	// ProcessEventHandlerの作成
	processEventHandler := &ProcessEventHandler{
		memoryReader: memoryReader,
		processName:  addressResolver.GetProcessName(),
	}

	// ProcessMonitorの初期化
	monitorConfig := &process.MonitorConfig{
		MonitorInterval: 1 * time.Second,
		EventHandler:    processEventHandler,
	}
	app.processMonitor = process.NewProcessMonitor(monitorConfig, processFinder)

	// HTTPサーバーの設定
	if err := app.setupHTTPServer(); err != nil {
		return nil, fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	return app, nil
}

// NewApplicationWithCallback creates and initializes a new Application with process status callback
// 関心の分離の原則に従い、GUI用のコールバック付きアプリケーション生成
func NewApplicationWithCallback(flags *config.Flags, callback ProcessStatusCallback) (*Application, error) {
	// Context設定
	ctx, cancel := context.WithCancel(context.Background())

	app := &Application{
		flags:  flags,
		ctx:    ctx,
		cancel: cancel,
	}

	// AddressResolverの作成（組み込み設定から直接作成）
	addressResolver, err := version.NewConfigAddressResolverByVersion(flags.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to create address resolver: %w", err)
	}

	// ProcessFinderの初期化
	processFinder := process.NewWindowsProcessFinder()

	// MemoryReaderの初期化
	memoryReader := memory.NewWindowsMemoryReader(processFinder)
	app.memoryReader = memoryReader

	// Collectorの初期化
	basicCollector := collector.NewBasicCollector(addressResolver, memoryReader)
	app.collector = basicCollector

	// Pollerの初期化
	app.poller = collector.NewPoller(basicCollector)

	// ProcessEventHandlerの作成（コールバック付き）
	processEventHandler := &ProcessEventHandler{
		memoryReader: memoryReader,
		processName:  addressResolver.GetProcessName(),
		callback:     callback,
	}

	// ProcessMonitorの初期化
	monitorConfig := &process.MonitorConfig{
		MonitorInterval: 1 * time.Second,
		EventHandler:    processEventHandler,
	}
	app.processMonitor = process.NewProcessMonitor(monitorConfig, processFinder)

	// HTTPサーバーの設定
	if err := app.setupHTTPServer(); err != nil {
		return nil, fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	return app, nil
}

// setupHTTPServer configures the HTTP server with all endpoints
// カプセル化の原則に従い、HTTPサーバー設定の責任を分離
func (app *Application) setupHTTPServer() error {
	mux := http.NewServeMux()

	// Prometheusメトリクスエンドポイント
	prometheusExporter := exporter.NewPrometheusExporter(app.collector)
	mux.Handle("/metrics", prometheusExporter)

	// ヘルスチェックエンドポイント
	healthHandler := exporter.NewHealthHandler(app.processMonitor)
	mux.Handle("/health", healthHandler)

	// HTTPサーバーの作成
	app.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", app.flags.Port),
		Handler: mux,
	}

	return nil
}

// Start starts all application components
// 関心の分離の原則に従い、各コンポーネントの起動を管理
func (app *Application) Start() error {
	log.Println("YSO Exporter - Ys Origin Metrics Exporter for Prometheus")
	log.Println("Version: 0.0.1-dev")
	log.Printf("Configuration:")
	log.Printf("  Game Version: %s", app.flags.Version)
	log.Printf("  HTTP Port: %d", app.flags.Port)
	log.Printf("  Polling Interval: %v", app.flags.Interval)

	// AddressResolverからプロセス名を取得
	resolver, err := version.NewConfigAddressResolverByVersion(app.flags.Version)
	if err != nil {
		return fmt.Errorf("failed to get address resolver: %w", err)
	}
	processName := resolver.GetProcessName()

	log.Printf("Starting YSO Exporter on port %d...", app.flags.Port)
	log.Printf("Monitoring game version: %s", app.flags.Version)
	log.Printf("Target process: %s", processName)
	log.Printf("Polling interval: %v", app.flags.Interval)

	// ProcessMonitorの開始
	if err := app.processMonitor.Start(app.ctx, processName); err != nil {
		return fmt.Errorf("failed to start process monitor: %w", err)
	}
	log.Printf("Process monitor started for: %s", processName)

	// Pollerの開始
	if err := app.poller.Start(app.ctx, app.flags.Interval); err != nil {
		return fmt.Errorf("failed to start poller: %w", err)
	}
	log.Printf("Metrics poller started with interval: %v", app.flags.Interval)

	// HTTPサーバーの開始
	go func() {
		log.Printf("HTTP server starting on port %d", app.flags.Port)
		if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Println("Application started successfully")
	log.Printf("Endpoints available:")
	log.Printf("  Metrics: http://localhost:%d/metrics", app.flags.Port)
	log.Printf("  Health: http://localhost:%d/health", app.flags.Port)

	return nil
}

// WaitForShutdown waits for shutdown signals and performs graceful shutdown
// グレースフルシャットダウンの責任を分離
func (app *Application) WaitForShutdown() {
	// シグナルチャンネルの設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// シャットダウンシグナルの待機
	<-sigChan
	log.Println("Shutdown signal received, starting graceful shutdown...")

	// グレースフルシャットダウンの実行
	app.shutdown()
}

// shutdown performs graceful shutdown of all components
// 関心の分離の原則に従い、シャットダウン処理を管理
func (app *Application) shutdown() {
	// コンテキストのキャンセル
	app.cancel()

	// HTTPサーバーのシャットダウン
	if app.httpServer != nil {
		log.Println("Shutting down HTTP server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := app.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		} else {
			log.Println("HTTP server shut down successfully")
		}
	}

	// Pollerの停止
	if app.poller != nil {
		log.Println("Stopping metrics poller...")
		app.poller.Stop()
		log.Println("Metrics poller stopped successfully")
	}

	// ProcessMonitorの停止
	if app.processMonitor != nil {
		log.Println("Stopping process monitor...")
		if err := app.processMonitor.Stop(); err != nil {
			log.Printf("Process monitor stop error: %v", err)
		} else {
			log.Println("Process monitor stopped successfully")
		}
	}

	// MemoryReaderのクローズ
	if app.memoryReader != nil {
		log.Println("Closing memory reader...")
		if err := app.memoryReader.Close(); err != nil {
			log.Printf("Memory reader close error: %v", err)
		} else {
			log.Println("Memory reader closed successfully")
		}
	}
}

// Cleanup performs final cleanup operations
// リソースクリーンアップの責任を分離
func (app *Application) Cleanup() {
	if app.cancel != nil {
		app.cancel()
	}
}
