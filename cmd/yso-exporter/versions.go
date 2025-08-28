package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/na2na/yso-exporter/internal/config"
)

func listAvailableVersions() {
	versions := config.GetAvailableVersions()

	fmt.Println("Available game versions:")
	for i, version := range versions {
		fmt.Printf("%d. %s\n", i+1, version)
	}
}

func selectVersionInteractively() string {
	versions := config.GetAvailableVersions()

	fmt.Println("Select a game version:")
	for i, version := range versions {
		fmt.Printf("%d. %s\n", i+1, version)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter choice (1-", len(versions), "): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading input: %v", err)
		return ""
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(versions) {
		log.Printf("Invalid choice: %s", input)
		return ""
	}

	selectedVersion := versions[choice-1]
	fmt.Printf("Selected version: %s\n", selectedVersion)
	return selectedVersion
}

func getAvailableVersions() []string {
	return config.GetAvailableVersions()
}
