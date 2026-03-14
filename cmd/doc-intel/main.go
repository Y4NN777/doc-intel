package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Doc-Intel v0.1.0")
	fmt.Println("Terminal-native document intelligence")
	fmt.Println()
	
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Entry point - enforces INV-07 (zero network calls)
	// TODO: Initialize TUI and wire all modules
	return fmt.Errorf("not implemented")
}
