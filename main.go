package main

import (
	"embed"
	"os"

	"github.com/averyfreeman/ansipath/internal/app"
)

// embeddedAssets keeps palettes and shell adapters in the single executable.
//
//go:embed palettes/*.json shell_helpers/*.sh
var embeddedAssets embed.FS

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, embeddedAssets))
}
