# Changelog

All notable changes to GMK (GoMicroKit) will be documented in this file.

## [Unreleased]

## [1.3.0] - 2026-02-02

### Added
- Short CLI command name `gmk` (previously `gomicrokit`) for easier typing
- `version` command to display version, git commit, build date, and Go version
- `--dry-run` flag for `new` command to preview project structure
- Skip notification when files already exist during service generation
- Proper error messages with context throughout the codebase
- Unit tests for generator package (25.7% coverage)
- Unicode support in service name validation
- Underscore support in service names

### Changed
- Repository template now includes full CRUD implementation with pagination
- Service template now includes business logic methods
- Handler template now includes real REST implementations with validation
- DTO template now includes Create/Update input structs
- Model template now includes sample Name field and TableName method
- Test template now includes actual JSON serialization tests
- Improved error handling - functions return errors instead of calling os.Exit()
- Better UX with "next steps" guide after project creation

### Fixed
- Duplicate `writeHistory` call in service generation
- Redundant empty check in RemoveService
- Time message inconsistency ("5 minutes" vs "1 minute")
- File handle leak in template rendering (defer in loop)
- Non-deterministic file generation order (now sorted)
- Module path incorrectly lowercased
- Duplicate `getGoModule` function consolidated
- Handler template using strings instead of errors for `errorResponse()`
- Model template missing `Name` field referenced by handler and DTO

### Removed
- Unsafe `must()` function that called os.Exit()

## [0.1.0] - Initial Release

### Added
- Interactive project creation with Bubble Tea TUI
- Service generation with `make:service` command
- Service removal with `remove:service` command
- Fiber framework support
- MySQL database support
- GORM ORM integration
- Docker support with Dockerfile generation
- Repository pattern architecture
- Service layer architecture
- Generation history tracking with `.gen_history.json`
