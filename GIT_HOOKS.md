# Git Hooks

This repository includes Git hooks to maintain code quality and prevent common mistakes.

## Pre-commit Hook

Location: `.git/hooks/pre-commit`

### Purpose
Prevents committing binary executable files (Mach-O and ELF executables) to the repository.

### What it checks
- Mach-O executables (macOS binaries)
- ELF executables (Linux binaries)  
- Files in `bin/` directory that are executable
- Files with binary extensions: `.exe`, `.dll`, `.so`, `.dylib`, `.a`, `.o`

### Installation
The hook is automatically installed when you clone this repository. If you need to reinstall it:

```bash
chmod +x .git/hooks/pre-commit
```

### Testing the hook
The hook will prevent commits containing binary files and provide helpful error messages with suggestions for resolution.

### Bypassing the hook (NOT recommended)
In rare cases where you need to commit a binary file (documentation, test fixtures, etc.), you can bypass the hook:

```bash
git commit --no-verify -m "commit message"
```

**Warning**: Only bypass the hook if you're absolutely certain the binary file should be in the repository.

### Build artifacts
Build artifacts should be placed in directories that are ignored by git:
- `bin/` - for compiled binaries
- `build/` - for build outputs  
- `dist/` - for distribution packages

These directories are already included in `.gitignore`.
