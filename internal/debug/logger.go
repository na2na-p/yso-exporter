package debug

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"
)

var (
	// Enabled indicates if debug logging is enabled
	Enabled bool
	
	// LogFile is the debug log file
	LogFile *os.File
)

// Init initializes debug logging
func Init(enabled bool) {
	Enabled = enabled
	
	if !enabled {
		return
	}
	
	// Create debug log file
	logFileName := fmt.Sprintf("yso-exporter-debug-%s.log", time.Now().Format("20060102-150405"))
	file, err := os.Create(logFileName)
	if err != nil {
		log.Printf("Failed to create debug log file: %v", err)
		return
	}
	
	LogFile = file
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	
	Log("Debug logging initialized to file: %s", logFileName)
	Log("===========================================")
}

// Close closes the debug log file
func Close() {
	if LogFile != nil {
		Log("===========================================")
		Log("Debug logging terminated")
		LogFile.Close()
	}
}

// Log writes a debug log message
func Log(format string, args ...interface{}) {
	if !Enabled {
		return
	}
	
	msg := fmt.Sprintf(format, args...)
	log.Output(2, "[DEBUG] "+msg)
	
	// Also print to console when debugging
	fmt.Printf("[DEBUG] %s\n", msg)
}

// LogError logs an error with stack trace
func LogError(err error, context string) {
	if !Enabled {
		return
	}
	
	if err == nil {
		return
	}
	
	Log("ERROR in %s: %v", context, err)
	Log("Stack trace:\n%s", debug.Stack())
}

// LogPanic logs panic information
func LogPanic(context string) {
	if !Enabled {
		return
	}
	
	if r := recover(); r != nil {
		Log("PANIC in %s: %v", context, r)
		Log("Stack trace:\n%s", debug.Stack())
		
		// Re-panic after logging
		panic(r)
	}
}

// Trace logs function entry/exit
type Trace struct {
	name string
	start time.Time
}

// Enter logs function entry
func Enter(name string) *Trace {
	if !Enabled {
		return &Trace{name: name, start: time.Now()}
	}
	
	t := &Trace{
		name:  name,
		start: time.Now(),
	}
	
	Log("ENTER: %s", name)
	return t
}

// Exit logs function exit
func (t *Trace) Exit() {
	if !Enabled {
		return
	}
	
	elapsed := time.Since(t.start)
	Log("EXIT:  %s (took %v)", t.name, elapsed)
}