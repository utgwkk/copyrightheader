// Package copyrightheader provides an analysis.Analyzer that verifies each Go
// source file begins with a copyright header comment.
package copyrightheader

import (
	"errors"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/tools/go/analysis"
)

// Config holds the analyzer configuration.
type Config struct {
	// Header is the canonical header text (comment markers stripped).
	// Used as the template when inserting/fixing. Required.
	Header string
	// PatternStart matches the first line of a valid header block.
	PatternStart string
	// PatternEnd matches the last line of a valid header block.
	PatternEnd string
	// PatternIgnore: if the file's first comment line matches, the file is skipped.
	PatternIgnore string
	// Exclude: file-path globs (doublestar semantics); matching files are skipped.
	Exclude []string
}

// Validate checks the configuration for consistency.
func (c Config) Validate() error {
	if c.Header == "" {
		return errors.New("copyrightheader: header must not be empty")
	}
	hasStart := c.PatternStart != ""
	hasEnd := c.PatternEnd != ""
	if hasStart != hasEnd {
		return errors.New("copyrightheader: pattern_start and pattern_end must both be set or both unset")
	}
	return nil
}

type analyzer struct {
	cfg         Config
	reStart     *regexp.Regexp
	reEnd       *regexp.Regexp
	reIgnore    *regexp.Regexp
	projectRoot string
	configPath  string
}

func (an *analyzer) run(pass *analysis.Pass) (any, error) {
	// If a config file was specified, load it (overrides flag values).
	if an.configPath != "" {
		cfg, err := LoadConfig(an.configPath)
		if err != nil {
			return nil, fmt.Errorf("copyrightheader: loading config: %w", err)
		}
		an.cfg = cfg
	}

	if err := an.cfg.Validate(); err != nil {
		return nil, err
	}

	if err := an.compileRegexes(); err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err == nil {
		an.projectRoot = filepath.Clean(cwd)
	}

	for _, file := range pass.Files {
		an.checkFile(pass, file)
	}
	return nil, nil
}

func (an *analyzer) compileRegexes() error {
	if an.cfg.PatternStart != "" {
		re, err := regexp.Compile(an.cfg.PatternStart)
		if err != nil {
			return fmt.Errorf("copyrightheader: invalid pattern_start: %w", err)
		}
		an.reStart = re
	}
	if an.cfg.PatternEnd != "" {
		re, err := regexp.Compile(an.cfg.PatternEnd)
		if err != nil {
			return fmt.Errorf("copyrightheader: invalid pattern_end: %w", err)
		}
		an.reEnd = re
	}
	if an.cfg.PatternIgnore != "" {
		re, err := regexp.Compile(an.cfg.PatternIgnore)
		if err != nil {
			return fmt.Errorf("copyrightheader: invalid pattern_ignore: %w", err)
		}
		an.reIgnore = re
	}
	return nil
}

func (an *analyzer) checkFile(pass *analysis.Pass, file *ast.File) {
	fileName := pass.Fset.File(file.Pos()).Name()

	if !an.isInProject(fileName) {
		return
	}
	if an.isExcluded(fileName) {
		return
	}

	cg := headerComment(file)

	if an.reIgnore != nil && cg != nil {
		if an.reIgnore.MatchString(firstCommentLine(cg)) {
			return
		}
	}

	if an.cfg.PatternStart != "" {
		an.checkFilePatternMode(pass, file, cg)
	} else {
		an.checkFileFallbackMode(pass, file, cg)
	}
}

func (an *analyzer) checkFilePatternMode(pass *analysis.Pass, file *ast.File, cg *ast.CommentGroup) {
	want := strings.TrimSpace(an.cfg.Header)

	if cg == nil {
		an.reportMissing(pass, file, want)
		return
	}

	if !an.reStart.MatchString(firstCommentLine(cg)) {
		an.reportMissing(pass, file, want)
		return
	}

	if an.hasEndLine(cg) {
		an.checkFormatting(pass, file, cg)
		return
	}

	an.reportIncomplete(pass, file, cg, want)
}

func (an *analyzer) checkFileFallbackMode(pass *analysis.Pass, file *ast.File, cg *ast.CommentGroup) {
	want := strings.TrimSpace(an.cfg.Header)

	if cg == nil {
		an.reportMissing(pass, file, want)
		return
	}

	if strings.TrimSpace(cg.Text()) == want {
		an.checkFormatting(pass, file, cg)
		return
	}

	an.reportMismatch(pass, file, want)
}

func (an *analyzer) reportMissing(pass *analysis.Pass, file *ast.File, want string) {
	pass.Report(analysis.Diagnostic{
		Pos:     file.Package,
		Message: "missing copyright header",
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "Insert copyright header",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     file.FileStart,
						End:     file.FileStart,
						NewText: []byte(renderHeaderComment(want) + "\n\n"),
					},
				},
			},
		},
	})
}

func (an *analyzer) reportMismatch(pass *analysis.Pass, file *ast.File, want string) {
	pass.Report(analysis.Diagnostic{
		Pos:     file.Package,
		Message: "copyright header does not match",
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "Insert the required copyright header",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     file.FileStart,
						End:     file.FileStart,
						NewText: []byte(renderHeaderComment(want) + "\n\n"),
					},
				},
			},
		},
	})
}

func (an *analyzer) reportIncomplete(pass *analysis.Pass, file *ast.File, cg *ast.CommentGroup, want string) {
	pass.Report(analysis.Diagnostic{
		Pos:     file.Package,
		Message: "incomplete copyright header",
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: "Replace with the complete copyright header",
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
}

func (an *analyzer) checkFormatting(pass *analysis.Pass, file *ast.File, cg *ast.CommentGroup) {
	want := strings.TrimSpace(an.cfg.Header)

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

func (an *analyzer) isInProject(path string) bool {
	if an.projectRoot == "" {
		return true
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return true
	}
	root := an.projectRoot
	if !filepath.IsAbs(root) {
		root, err = filepath.Abs(root)
		if err != nil {
			return true
		}
	}
	return abs == root || strings.HasPrefix(abs, root+string(filepath.Separator))
}

func (an *analyzer) isExcluded(path string) bool {
	if len(an.cfg.Exclude) == 0 {
		return false
	}
	rel := path
	if an.projectRoot != "" {
		if r, err := filepath.Rel(an.projectRoot, path); err == nil {
			rel = r
		}
	}
	rel = filepath.ToSlash(rel)
	for _, pattern := range an.cfg.Exclude {
		if matched, _ := doublestar.Match(pattern, rel); matched {
			return true
		}
	}
	return false
}

func (an *analyzer) hasEndLine(cg *ast.CommentGroup) bool {
	for _, c := range cg.List {
		for _, line := range strings.Split(c.Text, "\n") {
			if an.reEnd.MatchString(line) {
				return true
			}
		}
	}
	return false
}

// stringSlice is a flag.Value that accumulates comma-separated values into a *[]string.
type stringSlice struct {
	dst *[]string
}

func (s *stringSlice) String() string {
	if s.dst == nil {
		return ""
	}
	return strings.Join(*s.dst, ",")
}

func (s *stringSlice) Set(v string) error {
	for _, part := range strings.Split(v, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			*s.dst = append(*s.dst, part)
		}
	}
	return nil
}

// New creates the copyrightheader analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	an := &analyzer{cfg: cfg}

	a := &analysis.Analyzer{
		Name: "copyrightheader",
		Doc:  "checks that every Go file starts with a copyright header comment",
		Run:  an.run,
	}
	a.Flags.StringVar(&an.cfg.Header, "header", cfg.Header,
		"required copyright header text (comment markers stripped, whitespace trimmed)")
	a.Flags.StringVar(&an.cfg.PatternStart, "pattern-start", cfg.PatternStart,
		"regex matching the first line of a valid header block")
	a.Flags.StringVar(&an.cfg.PatternEnd, "pattern-end", cfg.PatternEnd,
		"regex matching the last line of a valid header block")
	a.Flags.StringVar(&an.cfg.PatternIgnore, "pattern-ignore", cfg.PatternIgnore,
		"regex matching the first comment line; matching files are skipped")
	a.Flags.Var(&stringSlice{dst: &an.cfg.Exclude}, "exclude",
		"comma-separated file-path globs to skip (doublestar: ** crosses /)")
	a.Flags.StringVar(&an.configPath, "config", "",
		"path to YAML config file (overrides other flags)")

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

// headerComment returns the first non-directive comment group before the
// package clause. Directive-only groups (e.g. //go:build) are skipped because
// ast.CommentGroup.Text() strips directives, leaving an empty string.
func headerComment(file *ast.File) *ast.CommentGroup {
	for _, cg := range file.Comments {
		if cg.Pos() >= file.Package {
			break
		}
		if strings.TrimSpace(cg.Text()) != "" {
			return cg
		}
	}
	return nil
}

// firstCommentLine returns the raw text of the first line of the first comment
// in the group (including the // or /* prefix).
func firstCommentLine(cg *ast.CommentGroup) string {
	if len(cg.List) == 0 {
		return ""
	}
	text := cg.List[0].Text
	if strings.HasPrefix(text, "/*") {
		if idx := strings.IndexByte(text, '\n'); idx >= 0 {
			return text[:idx]
		}
		return text
	}
	return text
}
