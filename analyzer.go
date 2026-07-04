// Package copyrightheader provides an analysis.Analyzer that verifies each Go
// source file begins with a copyright header comment.
package copyrightheader

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Config configures the analyzer.
type Config struct {
	// Header is the exact text (comment markers stripped) that the leading
	// comment must contain. Both the file comment and this value are trimmed
	// before comparison.
	Header string
}

// New returns an analyzer that reports files whose leading comment does not
// exactly match cfg.Header (after trimming), or that have no leading comment
// at all, or that have no blank line between the copyright comment and the
// package clause.
func New(cfg Config) *analysis.Analyzer {
	// header is writable so that the -header flag can override cfg.Header when
	// the binary is invoked from the command line via singlechecker.
	header := cfg.Header

	a := &analysis.Analyzer{
		Name: "copyrightheader",
		Doc:  "checks that every Go file starts with a copyright header comment",
	}
	a.Flags.StringVar(&header, "header", cfg.Header,
		"required copyright header text (comment markers stripped, whitespace trimmed)")

	a.Run = func(pass *analysis.Pass) (any, error) {
		want := strings.TrimSpace(header)
		if want == "" {
			return nil, fmt.Errorf("copyrightheader: header must not be empty")
		}

		for _, file := range pass.Files {
			cg := headerComment(file)
			if cg == nil {
				pass.Reportf(file.FileStart, "missing copyright header comment")
				continue
			}
			if got := strings.TrimSpace(cg.Text()); got != want {
				pass.Reportf(file.Package, "copyright header does not match the required text")
				continue
			}
			// file.Doc points to the comment group that is attached to the
			// package clause (i.e. immediately before it with no blank line).
			// If the copyright comment has no blank line separating it from
			// the package clause, the Go parser treats it as the package doc
			// and file.Doc == cg.
			if file.Doc == cg {
				pass.Reportf(file.Package,
					"copyright header must be separated from the package clause by a blank line")
			}
		}
		return nil, nil
	}

	return a
}

// headerComment returns the first "real" comment group that appears before the
// package clause in file. It skips comment groups whose Text() is empty after
// trimming — this naturally excludes directive-only groups (e.g. //go:build)
// because ast.CommentGroup.Text() strips directives from its output.
func headerComment(file *ast.File) *ast.CommentGroup {
	for _, cg := range file.Comments {
		if cg.Pos() >= file.Package {
			// Comment groups are stored in position order; once we pass the
			// package keyword there is nothing more to check.
			break
		}
		if strings.TrimSpace(cg.Text()) != "" {
			return cg
		}
	}
	return nil
}

