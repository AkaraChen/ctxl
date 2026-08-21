package main

import (
	"fmt"
	"os"

	"github.com/AkaraChen/ctxl/cli"
)

func main() {
	if err := cli.New(cli.Options{Name: "ctxl"}).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ctxl: %v\n", err)
		os.Exit(1)
	}
}
