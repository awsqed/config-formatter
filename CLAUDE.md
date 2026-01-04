# AGENT INSTRUCTIONS

**Last Updated:** January 2026 | **For:** config-formatter - YAML Configuration Formatter

## Table of Contents
- [Initial Setup](#initial-setup)
- [Development Workflow](#development-workflow)
- [Quick Reference](#quick-reference)
- [Code Quality Standards](#code-quality-standards)
- [Style Guidelines](#style-guidelines)
- [Project Context](#project-context)
  - [Overview](#overview)
  - [Architecture](#architecture)
  - [Key Files](#key-files)
  - [Formatter Interface](#formatter-interface)
  - [Context-Aware Key Ordering](#context-aware-key-ordering)
  - [Formatter Implementations](#formatter-implementations)
  - [Testing](#testing)
- [Adding New Formatters](#adding-new-formatters)
- [Common Pitfalls](#common-pitfalls)
- [Troubleshooting](#troubleshooting)
- [Maintaining This Document](#maintaining-this-document)

---

The following conventions must be followed for any changes in this repository.

## Initial Setup

1. Clone the repository and navigate to the project directory.
2. Ensure Go 1.25.5 or later is installed (check with `go version`).
3. Download dependencies: `make mod-download` or `go mod download`.
4. Verify the build works: `make build`.

## Development Workflow

1. Make changes to the codebase.
2. Test locally with sample files: `go run main.go -input compose.yml`.
3. Run tests (when implemented): `make test`.
4. Build for current platform: `make build`.
5. Keep the working tree clean before finishing.

## Quick Reference

| Task | Command |
|------|---------|
| Run without building | `go run main.go -input <file>` |
| Run with Makefile | `make run FILE=<file>` |
| Build current platform | `make build` |
| Build all platforms | `make build-all` |
| Build specific platform | `make build-linux-amd64`, `make build-darwin-arm64`, etc. |
| Run tests | `make test` |
| Clean build artifacts | `make clean` |
| Update dependencies | `make mod-tidy` |

## Code Quality Standards

- Follow idiomatic Go practices.
- Handle errors explicitly; wrap errors with context using `fmt.Errorf("context: %w", err)`.
- Maintain test coverage; add tests for new functionality (none currently exist).
- Use table-driven tests where appropriate.

## Style Guidelines

- Follow Go conventions and clean architecture practices.
- Prefer small, focused interfaces and dependency injection via constructors.
- Document all exported identifiers with comprehensive GoDoc comments.
- Use interface-driven development; public functions should accept interfaces, not concrete types.
- Handle errors explicitly with context:

  **Example:**
  ```go
  // Good
  if err := formatter.Format(data, indent); err != nil {
      return fmt.Errorf("failed to format %s config: %w", name, err)
  }

  // Bad - no context
  if err := formatter.Format(data, indent); err != nil {
      return err
  }
  ```

---

## Project Context

### Overview

**config-formatter** is a modular CLI tool for standardizing YAML configuration files with consistent indentation and format-specific directive ordering. It currently supports Docker Compose and Traefik configurations with an extensible plugin architecture for adding more formats.

### Architecture

#### Core Design: Plugin-Based Formatters

The project uses a modular plugin architecture where each config format (Docker Compose, Traefik, etc.) is an independent formatter module that implements the `Formatter` interface.

#### Key Packages

- **`formatter/`** - Core interface and base utilities
  - `formatter.go` - Formatter interface definition and BaseFormatter helper
  - `BaseFormatter.FormatYAML()` - Shared YAML processing logic used by all formatters

- **`modules/dockercompose/`** - Docker Compose formatter implementation
  - Handles docker-compose.yml files with service-specific ordering
  - Special processing: environment normalization, port quoting, service spacing

- **`modules/traefik/`** - Traefik formatter implementation
  - Handles traefik configuration with protocol-level organization
  - Context-aware ordering for nested sections (http/tcp/udp)

- **`main.go`** - CLI entry point and formatter registry
  - Flag parsing: `-input`, `-output`, `-w`, `-indent`, `-check`, `-type`
  - Formatter selection: auto-detection or explicit type
  - Output handling: stdout, file, or in-place modification

#### Design Patterns

- **Plugin Architecture**: Formatters register in main.go's `formatters` slice
- **Interface-Driven Design**: All formatters implement the same 3-method interface
- **Template Method Pattern**: BaseFormatter provides shared YAML processing via `FormatYAML()`
- **Strategy Pattern**: Different ordering strategies per context level (top-level, service-level, etc.)

### Key Files

- **`main.go`** - CLI entry point, flag parsing, formatter selection and orchestration
- **`formatter/formatter.go`** - Core interface definition and BaseFormatter helper
- **`modules/dockercompose/dockercompose.go`** - Docker Compose formatter (~400 lines)
- **`modules/traefik/traefik.go`** - Traefik formatter (~500 lines)
- **`Makefile`** - Build automation with cross-platform support

### Formatter Interface

All formatters must implement these three methods:

```go
type Formatter interface {
    // Format takes raw YAML data and returns formatted output
    Format(data []byte, indent int) ([]byte, error)

    // Name returns the formatter identifier (e.g., "docker-compose", "traefik")
    Name() string

    // CanHandle determines if this formatter can process the given file
    CanHandle(filename string, data []byte) bool
}
```

**BaseFormatter Helper**: Provides `FormatYAML()` which all formatters use to:
- Parse YAML into `yaml.Node` structure
- Apply context-specific key ordering during node traversal
- Handle comment preservation automatically
- Clean up empty lines in final output
- Encode back to YAML with specified indentation

### Context-Aware Key Ordering

**Critical Pattern**: The same key can appear at different nesting levels with different semantic meanings. Formatters must apply different ordering rules based on context.

**Example in Traefik:**
- `middlewares` at `http` level = middleware definitions (ordered with routers/services)
- `middlewares` at `http.routers.X` level = middleware references (ordered with router config)

**Example in Docker Compose:**
- `volumes` at top-level = volume definitions (ordered with networks/configs)
- `volumes` at service-level = volume mounts (ordered with ports/environment)

**Implementation:**
Each formatter defines multiple ordering maps (top-level, service-level, protocol-level, etc.) and passes the appropriate map to `BaseFormatter.FormatYAML()` based on the current context during YAML node traversal.

### Formatter Implementations

#### Docker Compose (`modules/dockercompose/dockercompose.go`)

- **Detection Logic**:
  - Filename patterns: `docker-compose`, `compose.yml`, `compose.yaml`
  - Content detection: presence of top-level keys `services` or `version`

- **Ordering Philosophy**: Metadata → Infrastructure → Services
  - Top-level: `version`, `name`, `networks`, `volumes`, `configs`, `secrets`, `services`
  - Service-level: Container identity → Execution → Configuration → Connectivity → Dependencies

- **Special Processing**:
  - **Environment normalization**: Converts array format to map with smart quoting (preserves numbers, booleans; quotes strings)
  - **Port normalization**: Ensures all port mappings are quoted strings (prevents YAML parser issues with `80:80` vs `8080:80`)
  - **Service spacing**: Adds blank lines between service entries for readability

- **Context Levels**:
  - Top-level context (infrastructure directives)
  - Service-level context (per-service configuration)

#### Traefik (`modules/traefik/traefik.go`)

- **Detection Logic**:
  - Filename pattern: `traefik`
  - Content detection: presence of top-level keys `http`, `tcp`, `udp`, `entryPoints`, `providers`, `certificatesResolvers`, `api`

- **Ordering Philosophy**: Global config → EntryPoints → Providers → Certificates → Protocols
  - Top-level: Global settings, `entryPoints`, `providers`, `certificatesResolvers`, then protocol sections
  - Protocol-level (http/tcp/udp): `routers`, `services`, `middlewares`, `serversTransports`
  - Router-level: Rule/priority → Service → Middlewares → TLS
  - 15+ middleware types organized by function (path modification, filtering, auth, etc.)

- **Special Handling**:
  - Context-aware ordering prevents key collisions
  - Protocol sections (http/tcp/udp) get distinct ordering from top-level
  - Middleware definitions vs references handled separately

- **Context Levels**:
  - Top-level context (global configuration)
  - Protocol-level context (http/tcp/udp sections)
  - Router-level context (individual router configuration)
  - Service-level context (individual service configuration)

### Testing

**Current State**: No tests currently implemented (no `*_test.go` files exist).

**Testing Framework Ready**: `make test` target exists in Makefile, runs `go test -v ./...`.

**Recommended Test Coverage** (when implementing):
- **Key Ordering**: Verify correct ordering at each context level
- **Special Processing**: Environment/port normalization, service spacing
- **Auto-Detection**: Filename patterns and content inspection for both formatters
- **Edge Cases**: Malformed YAML, empty files, comment preservation, deeply nested structures
- **Output Modes**: Stdout, file write, in-place modification (`-w`), check mode (`-check`)
- **Indentation**: Custom indent values (default 2, test with 4, 6, etc.)

---

## Adding New Formatters

To add support for a new config format:

### Step-by-Step Guide

1. **Create Module Directory**: `modules/yourformat/`

2. **Implement Formatter Interface** in `modules/yourformat/yourformat.go`:

```go
package yourformat

import "config-formatter/formatter"

type YourFormatter struct {
    formatter.BaseFormatter
}

func New() *YourFormatter {
    return &YourFormatter{}
}

func (f *YourFormatter) Name() string {
    return "yourformat"
}

func (f *YourFormatter) CanHandle(filename string, data []byte) bool {
    // Check filename patterns
    if strings.Contains(filename, "yourformat") {
        return true
    }

    // Check content for known top-level keys
    var content map[string]interface{}
    if err := yaml.Unmarshal(data, &content); err != nil {
        return false
    }

    // Return true if known keys exist
    _, hasKey1 := content["known_key"]
    _, hasKey2 := content["another_key"]
    return hasKey1 || hasKey2
}

func (f *YourFormatter) Format(data []byte, indent int) ([]byte, error) {
    // Define ordering for top-level keys
    topLevelOrder := map[string]int{
        "key1": 0,
        "key2": 1,
        "key3": 2,
    }

    // Use BaseFormatter's FormatYAML helper
    return f.FormatYAML(data, indent, topLevelOrder)
}
```

3. **Define Ordering Logic**: Create ordering maps for each context level your format needs

4. **Register Formatter** in `main.go`:

```go
formatters := []formatter.Formatter{
    dockercompose.New(),
    traefik.New(),
    yourformat.New(),  // Add your formatter here
}
```

5. **Test Manually**:
```bash
go run main.go -input sample.yml -type yourformat
```

6. **Update Documentation**: Add formatter details to README.md and this file

### Implementation Tips

- Start with simple top-level ordering, add context-aware ordering as needed
- Use `BaseFormatter.FormatYAML()` for standard YAML processing
- Implement `CanHandle()` with both filename and content detection for robustness
- Study existing formatters (Docker Compose for special processing, Traefik for context-aware ordering)
- Test with real config files from your target format

---

## Common Pitfalls

- **Context confusion**: Remember that keys can appear at multiple nesting levels with different meanings—always use context-appropriate ordering maps
- **Auto-detection overlap**: Ensure `CanHandle()` is specific enough to avoid false positives (e.g., don't just check for common keys like `name` or `version`)
- **Comment preservation**: Comments are automatically preserved by `gopkg.in/yaml.v3`, but complex manipulations may lose them—use `BaseFormatter.FormatYAML()` to maintain comments
- **YAML type conversions**: Ports like `80:80` are parsed as integers (base-60!), quote them as strings to preserve format
- **Empty line cleanup**: The `cleanEmptyLines()` helper removes excessive blank lines, but don't rely on it for structural spacing—add intentional spacing in Format() logic
- **Ordering map gaps**: Keys not in ordering maps are sorted alphabetically after ordered keys—this is intentional for extensibility
- **Indentation consistency**: Always use the provided `indent` parameter, don't hardcode spacing

---

## Troubleshooting

### Formatter Not Auto-Detected

**Symptoms:** Running with `-input file.yml` doesn't select your formatter, defaults to first registered

**Root cause:** `CanHandle()` returning false for the file

**Fix:**
1. Check filename pattern matching: `strings.Contains(filename, "yourformat")`
2. Verify content detection: add debug logging to see what keys are found
3. Test with explicit type: `go run main.go -input file.yml -type yourformat`
4. Ensure your formatter is registered in `main.go` formatters slice

### Keys Appearing in Wrong Order

**Symptoms:** Some keys are alphabetically sorted when they should follow specific order

**Root cause:** Keys missing from ordering map

**Fix:**
1. Check your ordering map includes all keys you want to order
2. Remember: unmapped keys are automatically sorted alphabetically after mapped keys
3. For context-aware ordering, ensure you're passing the correct ordering map for the current nesting level
4. Add debug logging in `BaseFormatter.FormatYAML()` to see which ordering map is being used

### Comments Disappearing

**Symptoms:** YAML comments are lost after formatting

**Root cause:** Using `yaml.v2` or manual string manipulation instead of `yaml.v3` node API

**Fix:**
1. Ensure you're using `gopkg.in/yaml.v3` (check go.mod)
2. Use `BaseFormatter.FormatYAML()` which preserves comments via node-level processing
3. Avoid string manipulation or regex on YAML content—work with parsed nodes
4. Test with sample files containing comments to verify preservation

### Port Numbers Changing Format

**Symptoms:** Port mapping `80:80` becomes `4800` or other unexpected values

**Root cause:** YAML parser interprets colon-separated numbers as base-60 (sexagesimal)

**Fix:**
1. Quote port strings: `"80:80"` instead of `80:80`
2. For Docker Compose formatter: port normalization already handles this in `normalizeValue()`
3. For new formatters: add similar normalization logic for port-like strings
4. Always test with common ports (80:80, 443:443, 8080:80)

### Build Fails on Cross-Platform

**Symptoms:** `make build-all` or specific platform builds fail

**Root cause:** Platform-specific code or missing GOOS/GOARCH combinations

**Fix:**
1. Verify Go version supports target platform: `go tool dist list`
2. Check for platform-specific imports or build tags
3. This project should be platform-agnostic (no CGO, no OS-specific packages)
4. Review Makefile platform targets for typos in GOOS/GOARCH values

---

## Maintaining This Document

**When to Update:** After completing significant work on the codebase, update this document with lessons learned and patterns discovered.

**What to Add:**
- New formatter implementations and their unique patterns
- Common issues discovered during development or testing
- Architectural decisions and their rationale (e.g., why context-aware ordering was chosen)
- Troubleshooting steps for new classes of issues
- Updates to development workflows or tooling

**Format:**
- Add concrete, actionable information
- Include code examples where helpful
- Keep entries concise but complete
- Update the "Last Updated" date at the top
- Update the Table of Contents if adding new sections

**For Future Agents:**
After completing your work session, review your changes and add relevant learnings to the appropriate sections above. This ensures institutional knowledge is preserved and future development is more efficient.
