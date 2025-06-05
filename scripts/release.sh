#!/bin/bash

# GoIRC Release Script
# Usage: ./scripts/release.sh <version>
# Example: ./scripts/release.sh v1.0.0

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION="$1"

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format vX.Y.Z (e.g., v1.0.0)"
    exit 1
fi

echo "🚀 Preparing release $VERSION"

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "Warning: You're not on the main branch (currently on $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check for uncommitted changes
if ! git diff --quiet; then
    echo "Error: You have uncommitted changes. Please commit or stash them first."
    exit 1
fi

# Check if tag already exists
if git tag | grep -q "^$VERSION$"; then
    echo "Error: Tag $VERSION already exists"
    exit 1
fi

# Run tests
echo "🧪 Running tests..."
go test -v ./...

# Build for all platforms locally (optional verification)
echo "🔨 Building for all platforms..."
make build-all

echo "✅ All builds successful"

# Create and push tag
echo "🏷️  Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

echo "📤 Pushing tag to origin..."
git push origin "$VERSION"

echo ""
echo "🎉 Release $VERSION has been triggered!"
echo ""
echo "The GitHub Actions workflow will now:"
echo "  - Build binaries for Linux, macOS, and Windows"
echo "  - Create a pre-release on GitHub"
echo "  - Upload all binary archives"
echo ""
echo "You can monitor the progress at:"
echo "https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^/]*\)\.git/\1/')/actions"
echo ""
echo "Once the workflow completes, you can find the release at:"
echo "https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^/]*\)\.git/\1/')/releases"

# Clean up local build artifacts
echo "🧹 Cleaning up local build artifacts..."
make clean
