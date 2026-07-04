# copyrightheader

A Go static analysis tool that checks every `.go` file starts with a copyright header comment.

## Features

- Reports files that are missing a copyright header
- Reports files whose copyright header does not match the required text
- Reports files where the copyright header is not separated from the `package` clause by a blank line
- Provides suggested fixes for all reported issues (auto-fixable via `go fix` / golangci-lint `--fix`)
- Requires the copyright header to appear before any `//go:build` build constraints (i.e. as the very first comment in the file); files with the wrong order are reported and an auto-fix that reorders them is provided
- Supports multi-line copyright headers

## Installation

```sh
go install github.com/utgwkk/copyrightheader/cmd/copyrightheader@latest
```

## Usage

### Standalone

```sh
copyrightheader -header='Copyright 2026 Your Name' ./...
```

### As a `go vet` tool

```sh
go vet -vettool=$(which copyrightheader) -copyrightheader.header='Copyright 2026 Your Name' ./...
```

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
          header: 'Copyright 2026 Your Name'
```

> **Note:** golangci-lint uses Viper for config loading, which lowercases map keys. The settings key must be lowercase `header`.

### Multi-line header

Both the standalone flag and the golangci-lint plugin support multi-line headers. Provide the full text without comment markers; the tool adds `// ` prefixes automatically.

```sh
copyrightheader -header=$'MIT License\n\nCopyright (c) 2026 Your Name' ./...
```
