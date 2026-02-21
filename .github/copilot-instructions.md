# Project Guidelines

## Code Style

- **Go 1.24.11** - Follow standard Go formatting (gofmt)
- **Error handling**: Wrap errors with `fmt.Errorf` and `%w` for error chains (see [build.go](../internal/build/build.go#L76))
- **Named return values**: Use for deferred error logging pattern (see [build.go](../internal/build/build.go#L64-L72))
- **Resource cleanup**: Always defer cleanup with `_` for ignored errors: `defer func() { _ = file.Close() }()`
- **Package logger**: Declare `var l = logger.DefaultLogger` at package level
- **Test helpers**: Prefix with `reset` and use deferred cleanup (see [build_test.go](../internal/build/build_test.go#L17-L19))

## Architecture

**Command Structure:**

- CLI built with `spf13/cobra` - root command in [root.go](../cmd/root.go)
- Commands in `cmd/`, business logic in `internal/`
- Centralized flags via [cmdargs/root.go](../internal/cmdargs/root.go)
- `PersistentPreRunE` hook handles logging/config setup

**Key Internal Packages:**

- `internal/build/` - Build orchestration and packaging
- `internal/upload/` - Platform-specific uploaders (CurseForge, WoWInterface, Wago.io)
- `internal/repo/` - VCS abstraction (Git/SVN via [VcsRepo interface](../internal/repo/repo.go#L13-L23))
- `internal/toc/` - WoW TOC file parsing and game version management
- `internal/tokens/` - Token replacement system (`@project-version@`, `@build-date@`, etc.)
- `internal/logger/` - Custom logging with emoji support and hierarchy (VERBOSE → DEBUG → INFO → WARN → ERROR)

**Interface-Based Design:**

- Always create interfaces for testability (see [repo.go](../internal/repo/repo.go), [interface.go](../internal/github/interface.go))
- Provide mock implementations for testing (see [mock_repo.go](../internal/repo/mock_repo.go))

**Data Flow:**

1. CLI command → `Build()` in [build.go](../internal/build/build.go)
2. Parse TOC files → Load pkgmeta.yml → Prepare VCS repo
3. Fetch externals (parallel with WaitGroups) → Copy files → Token injection → Zip → Upload

## Build and Test

**Build Commands:**

```bash
make build      # Build to dist/
make tools      # Install dev dependencies
make test       # Run tests with coverage (uses scripts/test.sh)
go test ./...   # Run all tests
```

**Testing Patterns:**

- Use `github.com/stretchr/testify/assert` and `require`
- E2E tests in `internal/build/test_e2e/` subdirectories
- Reset global state in tests with deferred cleanup functions
- Skip network tests in CI: `if os.Getenv("CI") != "" { t.Skip() }`

**Coverage Requirements:**

- Coverage threshold configured via `CC_THRESHOLD` environment variable
- HTML reports generated via gopogh
- Cyclomatic complexity checked via gocyclo

## Project Conventions

**Logging:**

- Sub-loggers with context: `logger.GetSubLog("EXT")`
- Emoji prefixes configurable (disable with `--no-emoji`)
- Use `l.Success()` for completion messages
- Performance timing: `l.Time(start, "operation complete")`

**Configuration (Viper):**

- Global config at `~/.wbt.yaml`
- Interactive prompts for first-time setup
- WoW installation paths per game flavor (Retail, Classic, Cata)

**Concurrency:**

- Use WaitGroups for parallel operations (see [pkgmeta.go](../internal/pkg/pkgmeta.go#L91-L95))
- Error channels for async operations
- File watching with `fsnotify/fsnotify`

**Build Parameters:**

- Centralized struct in [build.go](../internal/build/build.go#L17-L44)
- Feature flags via `Skip*` fields: `SkipCopy`, `SkipUpload`, `SkipZip`

**Environment-Aware Execution:**

- CI detection changes behavior (colors, auto-updates)
- GitHub Actions detection via `GITHUB_ACTIONS` env var

## Integration Points

**External APIs:**

- **CurseForge**: Requires `CF_API_KEY` env var (see [curse.go](../internal/upload/curse.go))
- **WoWInterface**: Requires `WOWI_API_TOKEN` (see [wowi.go](../internal/upload/wowi.go))
- **Wago.io**: Requires `WAGO_API_TOKEN` (see [wago.go](../internal/upload/wago.go))
- **GitHub**: Requires `GITHUB_OAUTH` token (see [github/api.go](../internal/github/api.go))

**VCS Integration:**

- Git operations via `go-git/go-git/v5` (see [git_repo.go](../internal/repo/git_repo.go))
- SVN externals support with caching via `.lastUpdated` markers
- Git externals with tag/branch/commit support

**Token System:**

- Comprehensive replacement: `@project-version@`, `@build-date@`, `@file-hash@`, etc.
- Build type tokens: `@alpha@`, `@retail@`, `@version-classic@`
- Injected into .lua, .xml, .toc, .md, .txt files (see [injector.go](../internal/injector/injector.go))

## Security

**Environment Variables:**

- Loaded from `.env` via [secrets.go](../internal/secrets/secrets.go)
- Required tokens: `GITHUB_OAUTH`, `CF_API_KEY`, `WOWI_API_TOKEN`, `WAGO_API_TOKEN`
- Early validation before API calls
- Never log sensitive tokens

**Best Practices:**

- Use environment variables exclusively for credentials
- Custom error types for missing tokens: `ErrNoCurseApiKey`, `ErrNoWagoApiKey`
- Proper file permissions: 0755 for dirs, 0644 for files
- GitHub Actions integration uses `secrets.*` properly
