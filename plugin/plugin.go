// Package plugin adapts the copyrightheader analyzer to golangci-lint's
// module plugin system.
//
// To use this plugin, add the following to your .custom-gcl.yml:
//
//	plugins:
//	  - module: github.com/utgwkk/copyrightheader
//	    import: github.com/utgwkk/copyrightheader/plugin
//
// And enable it in .golangci.yml:
//
//	linters:
//	  enable:
//	    - copyrightheader
//	  settings:
//	    custom:
//	      copyrightheader:
//	        type: module
//	        settings:
//	          header: 'Copyright (c) 2026 Stefan Daurer\nSPDX-License-Identifier: EPL-2.0'
//	          patternstart: '^//\s*Copyright.*'
//	          patternend: '^//\s*SPDX-License-Identifier.*'
//	          patternignore: '//\s*Code generated.*DO NOT EDIT.*'
//	          exclude:
//	            - 'mocks/**'
//	            - 'vendor/**'
//
// Note: golangci-lint uses Viper for config loading, which lowercases map
// keys. All settings keys must be lowercase.
package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/utgwkk/copyrightheader"
)

func init() {
	register.Plugin("copyrightheader", New)
}

// Settings maps to linters.settings.custom.copyrightheader.settings
// in .golangci.yml.
type Settings struct {
	Header        string   `json:"header"`
	PatternStart  string   `json:"patternstart"`
	PatternEnd    string   `json:"patternend"`
	PatternIgnore string   `json:"patternignore"`
	Exclude       []string `json:"exclude"`
}

type plugin struct {
	settings Settings
}

var _ register.LinterPlugin = (*plugin)(nil)

// New is the plugin constructor called by golangci-lint.
func New(raw any) (register.LinterPlugin, error) {
	settings, err := register.DecodeSettings[Settings](raw)
	if err != nil {
		return nil, err
	}
	cfg := copyrightheader.Config{
		Header:        settings.Header,
		PatternStart:  settings.PatternStart,
		PatternEnd:    settings.PatternEnd,
		PatternIgnore: settings.PatternIgnore,
		Exclude:       settings.Exclude,
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &plugin{settings: settings}, nil
}

// BuildAnalyzers returns the list of analyzers provided by this plugin.
func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	cfg := copyrightheader.Config{
		Header:        p.settings.Header,
		PatternStart:  p.settings.PatternStart,
		PatternEnd:    p.settings.PatternEnd,
		PatternIgnore: p.settings.PatternIgnore,
		Exclude:       p.settings.Exclude,
	}
	return []*analysis.Analyzer{
		copyrightheader.New(cfg),
	}, nil
}

// GetLoadMode returns LoadModeSyntax. We only inspect comments in the already-
// parsed AST, so no type information is needed.
func (p *plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}
