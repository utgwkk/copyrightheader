package copyrightheader_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/utgwkk/copyrightheader"
)

func TestAnalyzer(t *testing.T) {
	a := copyrightheader.New(copyrightheader.Config{
		Header: "Copyright 2026 utgwkk",
	})
	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), a, "a")
}

func TestAnalyzer_MultilineHeader(t *testing.T) {
	a := copyrightheader.New(copyrightheader.Config{
		Header: "Copyright 2026 utgwkk\nAll rights reserved.",
	})
	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), a, "b")
}
