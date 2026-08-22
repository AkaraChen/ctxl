package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/AkaraChen/ctxl/cli"
)

func main() {
	if err := cli.New(cli.Options{Name: "ctxl"}).Execute(); err != nil {
		if !errors.Is(err, cli.ErrNoMatch) {
			fmt.Fprintf(os.Stderr, "ctxl: %v\n", err)
		}
		os.Exit(1)
	}
}
