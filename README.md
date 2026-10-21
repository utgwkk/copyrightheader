# copyrightheader

A Go static analysis tool that checks every `.go` file starts with a copyright header comment.

## Features

- Reports files that are missing a copyright header
- Without patterns, reports files whose copyright header does not match the required text
- Pattern-based header detection: a header block is identified by start/end regexes, so existing non-header comments are never deleted
- Reports files where the copyright header is not separated from the `package` clause by a blank line
- Provides suggested fixes for all reported issues (auto-fixable via `go fix` / golangci-lint `--fix`)
- Requires the copyright header to appear before any `//go:build` build constraints (i.e. as the very first comment in the file); files with the wrong order are reported and an auto-fix that reorders them is provided
- Supports multi-line copyright headers
- Skips files matching `exclude` globs and files outside the project root (e.g. build-cache files)
- Configurable via CLI flags, a YAML config file, or golangci-lint settings

## Installation

```sh
go install github.com/utgwkk/copyrightheader/cmd/copyrightheader@latest
```

## Usage

### Standalone

```sh
copyrightheader -header='Copyright 2026 Your Name' ./...
```

With pattern-based detection (recommended, see [Configuration](#configuration)):

```sh
copyrightheader \
  -header=$'Copyright (c) 2026 Your Name\nSPDX-License-Identifier: MIT' \
  -pattern-start='^//\s*Copyright.*' \
  -pattern-end='^//\s*SPDX-License-Identifier.*' \
  ./...
```

### As a `go vet` tool

```sh
go vet -vettool=$(which copyrightheader) -copyrightheader.header='Copyright 2026 Your Name' ./...
```

Additional flags are passed the same way, e.g. `-copyrightheader.pattern-start='^//\s*Copyright.*'`.

### As a golangci-lint plugin

Add the plugin to `.custom-gcl.yml`:

```yaml
plugins:
  - module: github.com/utgwkk/copyrightheader
    import: github.com/utgwkk/copyrightheader/plugin
```

Enable and configure it in `.golangci.yml`:

```yaml
linters:
  enable:
    - copyrightheader
  settings:
    custom:
      copyrightheader:
        type: module
        settings:
          header: 'Copyright (c) 2026 Your Name\nSPDX-License-Identifier: MIT'
          patternstart: '^//\s*Copyright.*'
          patternend: '^//\s*SPDX-License-Identifier.*'
          patternignore: '//\s*Code generated.*DO NOT EDIT.*'
          exclude:
            - 'mocks/**'
            - 'vendor/**'
```

> **Note:** golangci-lint uses Viper for config loading, which lowercases map keys. All settings keys must be lowercase.

### Multi-line header

Both the standalone flag and the golangci-lint plugin support multi-line headers. Provide the full text without comment markers; the tool adds `// ` prefixes automatically.

```sh
copyrightheader -header=$'MIT License\n\nCopyright (c) 2026 Your Name' ./...
```

## Configuration

| Option | Standalone flag | Plugin key | Config key | Description |
|---|---|---|---|---|
| Header text (required) | `-header` | `header` | `header` | Canonical header text (comment markers stripped). Used as the template when inserting or fixing. |
| Header start pattern | `-pattern-start` | `patternstart` | `pattern_start` | Regex matching the first line of a valid header block. |
| Header end pattern | `-pattern-end` | `patternend` | `pattern_end` | Regex matching the last line of a valid header block. |
| Ignore pattern | `-pattern-ignore` | `patternignore` | `pattern_ignore` | Regex matching the first comment line; matching files are skipped (e.g. generated files). |
| Excluded paths | `-exclude` (comma-separated) | `exclude` (list) | `exclude` (list) | File-path globs to skip. Doublestar semantics: `**` crosses `/`. Matched against project-root-relative paths. |
| Config file | `-config` (standalone only) | — | — | Path to a YAML config file; overrides all other flags. |

`pattern_start` and `pattern_end` must both be set or both unset.

- **With patterns:** the header is a self-identifying block. A file is compliant when its first comment line matches `pattern_start` and the block contains a line matching `pattern_end`. An incomplete block is replaced; a missing block is prepended, so existing comments are never deleted.
- **Without patterns:** the first comment block must equal the header text exactly (trimmed, comment markers stripped). A mismatch is fixed by prepending the required header.

### YAML config file

```yaml
header: |
  Copyright (c) 2026 Your Name
  SPDX-License-Identifier: MIT
pattern_start: '^//\s*Copyright.*'
pattern_end: '^//\s*SPDX-License-Identifier.*'
pattern_ignore: '//\s*Code generated.*DO NOT EDIT.*'
exclude:
  - 'mocks/**'
  - 'vendor/**'
```

```sh
copyrightheader -config .copyrightheader.yml ./...
```
