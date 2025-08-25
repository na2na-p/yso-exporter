// +build !windows

package main

import "log"

func runGUI() {
	log.Fatal("GUI mode is only supported on Windows")
}