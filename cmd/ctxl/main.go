package main

import (
	"fmt"
	"os"

	"github.com/AkaraChen/ctxl/devcli"
)

func main() {
	if err := devcli.New().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ctxl: %v\n", err)
		os.Exit(1)
	}
}
