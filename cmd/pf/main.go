package main

import (
	"fmt"
	"os"

	"pf/internal/cli"
)

func main() {
	cmd := cli.NewRootCommand()
	
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}