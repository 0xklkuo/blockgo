package main

import (
	"fmt"
	"os"

	"blockgo/internal/version"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Println(version.String())
			return
		case "help", "--help", "-h":
			printUsage()
			return
		}
	}

	printUsage()
}

func printUsage() {
	fmt.Printf(`%s CLI

Status:
  Milestone 0 scaffold only. Functional commands arrive in later milestones.

Usage:
  blockgo <command>

Commands:
  version   Print version information
  help      Show this help

Planned commands:
  generate-key
  create-tx
  submit-tx
  show-chain
  show-utxo
`, version.Name)
}
