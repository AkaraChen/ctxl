package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/AkaraChen/ctxlayer/app"
	"github.com/AkaraChen/ctxlayer/schema"
)

//go:embed yoi.schema.json
var embedded []byte

func main() {
	s, err := schema.Parse(embedded)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ctxlayer: %v\n", err)
		os.Exit(1)
	}
	if err := app.New(app.Options{Name: "ctxlayer", Schema: s}).Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "ctxlayer: %v\n", err)
		os.Exit(1)
	}
}
