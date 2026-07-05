// Package copyrightheader provides an analysis.Analyzer that verifies each Go
// source file begins with a copyright header comment.
package copyrightheader

import (
	"errors"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type Config struct {
	// Header is the exact text (comment markers stripped) that the leading
	// comment must contain. Both the file comment and this value are trimmed
	// before comparison.
	Header string
}

type analyzer struct {
	// header is writable so that the -header flag can override the value from
	// Config when the binary is invoked from the command line via singlechecker.
	header string
}

func (an *analyzer) checkFile(pass *analysis.Pass, file *ast.File, want string) {
	cg := headerComment(file)
	if cg == nil {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Package,
			Message: "missing copyright header comment",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "Insert copyright header",
					TextEdits: []analysis.TextEdit{
						{
							// Insert at the very start of the file so that
							// the copyright header comes before any //go:build
							// directives.
							Pos:     file.FileStart,
							End:     file.FileStart,
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
	// The copyright header must be the very first comment in the file so
	// that build constraints (//go:build) come after it. If any comment
	// group precedes the header it can only be a directive-only group
	// (otherwise headerComment would have returned it instead) — i.e. a
	// misplaced build constraint.
	// file.Comments is guaranteed non-empty here: cg != nil means
	// headerComment found at least one element in file.Comments.
	if file.Comments[0] != cg {
		pass.Report(analysis.Diagnostic{
			Pos:     file.Package,
			Message: "copyright header must be the first comment in the file",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "Move the copyright header before the build constraints",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     file.FileStart,
							End:     file.FileStart,
							NewText: []byte(renderHeaderComment(want) + "\n\n"),
						},
						{
							// Remove the misplaced header and its trailing
							// blank line up to (but not including) the package
							// keyword.
							Pos:     cg.Pos(),
							End:     file.Package,
							NewText: nil,
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
		return nil, errors.New("copyrightheader: header must not be empty")
	}

	for _, file := range pass.Files {
		an.checkFile(pass, file, want)
	}
	return nil, nil
}

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

// headerComment returns the first comment group before the package clause.
// Directive-only groups (e.g. //go:build) are skipped because
// ast.CommentGroup.Text() strips directives, leaving an empty string.
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
