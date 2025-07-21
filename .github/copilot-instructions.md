# Copilot Instructions for Shared Go Libraries

This file provides repository-wide instructions for GitHub Copilot and other AI agents working on the sharedgolibs repository.

## Semantic Versioning

This repository follows [Semantic Versioning](https://semver.org/) strictly:

- **MAJOR** (x.0.0): Breaking changes that require code updates in consuming projects
- **MINOR** (0.x.0): New features that are backwards-compatible  
- **PATCH** (0.0.x): Bug fixes and backwards-compatible improvements

### Version Updating Rules

**🚨 CRITICAL**: Always update the version when making ANY functional change:

1. **Update Package Version Constant**: Update the `Version` constant in the modified package
2. **Update Documentation**: Reflect version changes in README.md examples  
3. **Update Tests**: Ensure version tests pass

### Package Version Format

Each package maintains its own version constant:

```go
// In pkg/util/env.go
const Version = "v0.1.0"

// In pkg/middleware/cors.go  
const Version = "v0.1.0"
```

### Examples of Version Changes:

#### MAJOR (Breaking Changes)
- Changing function signatures: `MustGetEnv(key, fallback)` → `MustGetEnv(config)`
- Removing public functions or packages
- Changing interface definitions
- Renaming packages or public types

#### MINOR (New Features)
- Adding new functions to existing packages
- Adding new packages (e.g., `pkg/database`, `pkg/cache`)
- Adding optional parameters with defaults
- Enhancing functionality without breaking existing code

#### PATCH (Bug Fixes)
- Fixing bugs in existing functions
- Improving error messages
- Performance improvements
- Documentation updates
- Test improvements

## Git Workflow

When making changes:

1. **Create Feature Branch**: `git checkout -b feature/description`
2. **Make Changes**: Implement the feature or fix
3. **Update Tests**: Ensure all tests pass and add new tests for new functionality
4. **Update Documentation**: Update README.md, function comments, and examples
5. **Determine Version**: Decide if change is MAJOR, MINOR, or PATCH
6. **Commit with Semantic Message**:
   ```
   feat(util): add new environment variable validation
   
   - Adds input validation to MustGetEnv function
   - Backwards compatible addition
   - Includes comprehensive tests
   - Updates package version to v0.1.1
   
   Package Version: util/v0.1.1 (PATCH - bug fix)
   ```
7. **Push Changes**: `git push origin feature/description`

## Package Design Principles

### Backwards Compatibility
- **NEVER** change existing function signatures without a MAJOR version bump
- **ALWAYS** add new optional parameters at the end with sensible defaults
- **PREFER** function options pattern for complex configurations

### Zero Dependencies
- **AVOID** external dependencies unless absolutely necessary
- **USE** only Go standard library when possible
- **JUSTIFY** any external dependencies in commit messages

### API Design
- **KEEP** APIs simple and focused
- **USE** clear, descriptive function names
- **PROVIDE** comprehensive examples in documentation
- **INCLUDE** error handling best practices

## Testing Requirements

### Coverage
- **MAINTAIN** 100% test coverage for all public functions
- **INCLUDE** edge cases and error conditions
- **TEST** both success and failure scenarios

### Test Structure
- **USE** table-driven tests for multiple scenarios
- **NAME** tests clearly: `TestFunctionName_Scenario`
- **INCLUDE** benchmarks for performance-critical functions

## Documentation Standards

### Function Documentation
```go
// MustGetEnv returns the value of the environment variable named by key.
// If the variable is not set or empty, returns the fallback value.
// This function unifies environment variable handling across projects.
//
// Example:
//   dbURL := util.MustGetEnv("DATABASE_URL", "localhost:5432")
func MustGetEnv(key, fallback string) string {
    // implementation
}
```

### Package Documentation
- **START** each package with a clear purpose statement
- **INCLUDE** usage examples in package comments
- **EXPLAIN** common use cases and patterns

## Integration Guidelines

### Consumer Projects
When updating sharedgolibs in allmytails or googleemu:

1. **TEST** thoroughly in development environment
2. **UPDATE** all import statements if package structure changes
3. **RUN** full test suites in consuming projects
4. **DOCUMENT** any required changes in consumer projects

### Release Process
1. **CREATE** release notes with breaking changes clearly marked
2. **UPDATE** consuming projects with specific package versions
3. **COMMUNICATE** breaking changes to all stakeholders
4. **DOCUMENT** version changes in package constants

## Automation and CI/CD

- **RUN** tests on every commit
- **VALIDATE** package version constants are updated when code changes
- **BLOCK** merges that break backwards compatibility without version bump
- **AUTOMATE** dependency updates in consuming projects

## Emergency Procedures

### Breaking Changes in PATCH/MINOR
If a breaking change is accidentally released in a PATCH or MINOR version:

1. **IMMEDIATELY** create a new MAJOR version with the breaking change
2. **REVERT** the breaking change in a new PATCH version
3. **COMMUNICATE** the issue to all consuming projects
4. **UPDATE** documentation with corrected version information

### Critical Bug Fixes
For security or critical bugs:

1. **FIX** the issue immediately
2. **CREATE** PATCH version even if other changes are pending
3. **NOTIFY** all consuming projects of the urgent update
4. **DOCUMENT** the severity and impact in release notes
