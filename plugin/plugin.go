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
//	          header: 'Copyright 2026 Your Name'
//
// Note: golangci-lint uses Viper for config loading, which lowercases map
// keys. The settings key must be lowercase "header".
package plugin

import (
	"errors"

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
	// Header is the exact copyright text that must appear at the top of every
	// .go file (comment markers stripped, whitespace trimmed).
	// The key must be lowercase because golangci-lint (Viper) lowercases keys.
	Header string `json:"header"`
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
	if settings.Header == "" {
		return nil, errors.New("copyrightheader: 'header' setting is required")
	}
	return &plugin{settings: settings}, nil
}

// BuildAnalyzers returns the list of analyzers provided by this plugin.
func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		copyrightheader.New(copyrightheader.Config{Header: p.settings.Header}),
	}, nil
}

// GetLoadMode returns LoadModeSyntax. We only inspect comments in the already-
// parsed AST, so no type information is needed. This is the key speed advantage
// over the bundled goheader linter.
func (p *plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}
