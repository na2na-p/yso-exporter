package config

import (
	"flag"
	"fmt"
	"os"
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
	
	flag.StringVar(&flags.Version, "version", "jp_1.1.0.3", "Game version to monitor")
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
	return []string{
		"jp_1.1.0.3",
		"steam_1.0.0",
		"gog_1.0.0",
		"jp_1.0.0",
	}
}

func PrintAvailableVersions() {
	fmt.Println("Available game versions:")
	for _, v := range GetAvailableVersions() {
		fmt.Printf("  - %s\n", v)
	}
}