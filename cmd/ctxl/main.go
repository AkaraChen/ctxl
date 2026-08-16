package main

import (
	"fmt"
	"os"

	"github.com/AkaraChen/ctxl/app"
)

func main() {
	if err := app.New(app.Options{Name: "ctxl"}).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ctxl: %v\n", err)
		os.Exit(1)
	}
}
