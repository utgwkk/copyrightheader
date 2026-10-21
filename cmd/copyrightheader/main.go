package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/utgwkk/copyrightheader"
)

func main() {
	singlechecker.Main(copyrightheader.New(copyrightheader.Config{}))
}
