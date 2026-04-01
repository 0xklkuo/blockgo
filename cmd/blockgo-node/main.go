package main

import (
	"flag"
	"fmt"

	"blockgo/internal/version"
)

func main() {
	var (
		configPath = flag.String("config", "./configs/node.example.json", "path to node config file")
		showVer    = flag.Bool("version", false, "print version information")
	)

	flag.Parse()

	if *showVer {
		fmt.Println(version.String())
		return
	}

	fmt.Printf("%s node scaffold ready\n", version.Name)
	fmt.Printf("config: %s\n", *configPath)
	fmt.Println("status: Milestone 0 scaffold only; node implementation starts in later milestones")
}
