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
		Header: `MIT License

Copyright (c) 2026 utagawa kiki

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`,
	})
	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), a, "b")
}

func TestAnalyzer_PatternMode(t *testing.T) {
	a := copyrightheader.New(copyrightheader.Config{
		Header:       "Copyright (c) 2026 utgwkk\nSPDX-License-Identifier: MIT",
		PatternStart: `^//\s*Copyright.*`,
		PatternEnd:   `^//\s*SPDX-License-Identifier.*`,
	})
	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), a, "c")
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     copyrightheader.Config
		wantErr bool
	}{
		{"valid minimal", copyrightheader.Config{Header: "x"}, false},
		{"valid pattern", copyrightheader.Config{Header: "x", PatternStart: "a", PatternEnd: "b"}, false},
		{"empty header", copyrightheader.Config{}, true},
		{"start only", copyrightheader.Config{Header: "x", PatternStart: "a"}, true},
		{"end only", copyrightheader.Config{Header: "x", PatternEnd: "b"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
