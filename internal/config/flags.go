package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Flags struct {
	Version      string
	Port         int
	Interval     time.Duration
	ListVersions bool
	Help         bool
}

func ParseFlags() *Flags {
	flags := &Flags{}

	flag.StringVar(&flags.Version, "version", "", "Game version to monitor (if not specified, will prompt for selection)")
	flag.IntVar(&flags.Port, "port", 8080, "HTTP server port for metrics endpoint")
	flag.DurationVar(&flags.Interval, "interval", 100*time.Millisecond, "Polling interval for memory reads")
	flag.BoolVar(&flags.ListVersions, "list-versions", false, "List all available game versions")
	flag.BoolVar(&flags.Help, "help", false, "Show help message")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "YSO Exporter - Ys Origin Metrics Exporter for Prometheus\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --version jp_1.1.0.3 --port 8080\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --list-versions\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --interval 50ms\n", os.Args[0])
	}

	flag.Parse()

	if flags.Help {
		flag.Usage()
		os.Exit(0)
	}

	return flags
}

func (f *Flags) Validate() error {
	if f.Port < 1 || f.Port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be between 1 and 65535)", f.Port)
	}

	if f.Interval < 10*time.Millisecond {
		return fmt.Errorf("interval too short: %v (minimum is 10ms)", f.Interval)
	}

	if f.Interval > 10*time.Second {
		return fmt.Errorf("interval too long: %v (maximum is 10s)", f.Interval)
	}

	availableVersions := GetAvailableVersions()
	if !f.ListVersions && f.Version != "" {
		found := false
		for _, v := range availableVersions {
			if v == f.Version {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unsupported version: %s\nAvailable versions: %v", f.Version, availableVersions)
		}
	}

	return nil
}

func GetAvailableVersions() []string {
	// configs/versionsディレクトリから利用可能なバージョンを動的に検出
	versionsDir := filepath.Join("configs", "versions")

	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		// ディレクトリが存在しない場合は空のリストを返す
		return []string{}
	}

	var versions []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			// .yaml拡張子を除いたファイル名をバージョン名として使用
			version := strings.TrimSuffix(entry.Name(), ".yaml")
			versions = append(versions, version)
		}
	}

	return versions
}

func PrintAvailableVersions() {
	fmt.Println("Available game versions:")
	for _, v := range GetAvailableVersions() {
		fmt.Printf("  - %s\n", v)
	}
}

func SelectVersionInteractive() string {
	versions := GetAvailableVersions()

	// バージョンが見つからない場合のエラー処理
	if len(versions) == 0 {
		fmt.Println("\nError: No version configuration files found in configs/versions/")
		fmt.Println("Please ensure at least one .yaml configuration file exists in the configs/versions/ directory.")
		os.Exit(1)
	}

	fmt.Println("\nYSO Exporter - Ys Origin Metrics Exporter")
	fmt.Println("\nAvailable game versions:")
	for i, v := range versions {
		fmt.Printf("  %d. %s\n", i+1, v)
	}

	// 1つしかバージョンがない場合は自動選択
	if len(versions) == 1 {
		fmt.Printf("\nAutomatically selecting the only available version: %s\n", versions[0])
		return versions[0]
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("\nSelect version (1-%d) or enter version name: ", len(versions))
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)

		// Check if input is a number
		if num, err := strconv.Atoi(input); err == nil {
			if num >= 1 && num <= len(versions) {
				selected := versions[num-1]
				fmt.Printf("Selected: %s\n", selected)
				return selected
			}
			fmt.Printf("Invalid selection. Please enter a number between 1 and %d.\n", len(versions))
			continue
		}

		// Check if input matches a version name
		for _, v := range versions {
			if v == input {
				fmt.Printf("Selected: %s\n", input)
				return input
			}
		}

		fmt.Printf("Invalid version: %s\n", input)
	}
}
