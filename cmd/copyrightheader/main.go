// Command copyrightheader runs the copyrightheader linter.
//
// It can be used in two ways:
//
//  1. Standalone:
//     copyrightheader -header='Copyright 2026 Your Name' ./...
//
//  2. As a vettool for go vet:
//     go vet -vettool=$(which copyrightheader) -copyrightheader.header='Copyright 2026 Your Name' ./...
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/utgwkk/copyrightheader"
)

func main() {
	// Config.Header is intentionally left empty here so that the -header flag
	// (registered inside copyrightheader.New) is the sole way to supply the
	// required header text when running from the command line.
	singlechecker.Main(copyrightheader.New(copyrightheader.Config{}))
}
