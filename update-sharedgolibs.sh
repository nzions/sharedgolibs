#!/bin/zsh

# Script to update github.com/nzions/sharedgolibs to the latest version
# This fixes common issues with Go module caching and proxy configuration

echo "🔧 Updating github.com/nzions/sharedgolibs to latest version..."

# Set GOPRIVATE to bypass proxy and use direct GitHub access
echo "📝 Setting GOPRIVATE for direct GitHub access..."
go env -w GOPRIVATE=github.com/nzions/*

# Clear the module cache to remove any stale versions
echo "🧹 Clearing Go module cache..."
go clean -modcache

# Remove any existing cached versions
echo "🗑️  Removing cached sharedgolibs..."
go mod download -x github.com/nzions/sharedgolibs 2>/dev/null || true

# Force update to the latest version
echo "⬇️  Fetching latest version from GitHub..."
go get -u github.com/nzions/sharedgolibs

# Clean up go.mod and go.sum
echo "🔄 Tidying go.mod..."
go mod tidy

# Verify the version
echo "✅ Current version:"
go list -m github.com/nzions/sharedgolibs

echo "🎉 Update complete! Your project now uses the latest sharedgolibs version."
echo ""
echo "💡 If you still have issues, make sure you have:"
echo "   - Git access to github.com/nzions/sharedgolibs"
echo "   - SSH key or GitHub token configured"
echo "   - No 'replace' directives for sharedgolibs in go.mod"
