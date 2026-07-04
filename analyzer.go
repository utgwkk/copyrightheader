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

// analyzer holds the state threaded through a single analysis pass.
type analyzer struct {
	// header is writable so that the -header flag can override the value from
	// Config when the binary is invoked from the command line via singlechecker.
	header string
}

func (an *analyzer) checkFile(pass *analysis.Pass, file *ast.File, want string) {
	cg := headerComment(file)
	if cg == nil {
		pass.Report(analysis.Diagnostic{
			Pos:     file.FileStart,
			Message: "missing copyright header comment",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "Insert copyright header",
					TextEdits: []analysis.TextEdit{
						{
							// Insert before the package clause so that any
							// preceding //go:build directives are left in place.
							Pos:     file.Package,
							End:     file.Package,
							NewText: []byte(renderHeaderComment(want) + "\n\n"),
						},
					},
				},
			},
		})
		return
	}
	if got := strings.TrimSpace(cg.Text()); got != want {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Package,
			Message: "copyright header does not match the required text",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "Replace with the required copyright header",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     cg.Pos(),
							End:     cg.End(),
							NewText: []byte(renderHeaderComment(want)),
						},
					},
				},
			},
		})
		return
	}
	// file.Doc points to the comment group that is attached to the
	// package clause (i.e. immediately before it with no blank line).
	// If the copyright comment has no blank line separating it from
	// the package clause, the Go parser treats it as the package doc
	// and file.Doc == cg.
	if file.Doc == cg {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Package,
			Message: "copyright header must be separated from the package clause by a blank line",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "Add a blank line after the copyright header",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     cg.End(),
							End:     cg.End(),
							NewText: []byte("\n"),
						},
					},
				},
			},
		})
	}
}

func (an *analyzer) run(pass *analysis.Pass) (any, error) {
	want := strings.TrimSpace(an.header)
	if want == "" {
		return nil, fmt.Errorf("copyrightheader: header must not be empty")
	}

	for _, file := range pass.Files {
		an.checkFile(pass, file, want)
	}
	return nil, nil
}

// New returns an analyzer that reports files whose leading comment does not
// exactly match cfg.Header (after trimming), or that have no leading comment
// at all, or that have no blank line between the copyright comment and the
// package clause.
func New(cfg Config) *analysis.Analyzer {
	an := &analyzer{header: cfg.Header}

	a := &analysis.Analyzer{
		Name: "copyrightheader",
		Doc:  "checks that every Go file starts with a copyright header comment",
		Run:  an.run,
	}
	a.Flags.StringVar(&an.header, "header", cfg.Header,
		"required copyright header text (comment markers stripped, whitespace trimmed)")

	return a
}

// renderHeaderComment converts plain header text into Go // comment form.
// Each line of text is prefixed with "// "; blank lines become bare "//".
// Multi-line headers are handled correctly.
func renderHeaderComment(text string) string {
	lines := strings.Split(text, "\n")
	var b strings.Builder
	for i, l := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		if l == "" {
			b.WriteString("//")
		} else {
			b.WriteString("// ")
			b.WriteString(l)
		}
	}
	return b.String()
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

