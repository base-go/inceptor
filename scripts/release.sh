#!/bin/bash
set -e

# Release script - builds all binaries and creates GitHub release
# Usage: ./scripts/release.sh [version]  - specify version like 1.0.1
#        ./scripts/release.sh            - auto-increment patch version

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

# Validate version format (x.x.x where x is a number)
validate_version() {
    if [[ ! "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        return 1
    fi
    return 0
}

# Normalize version to x.x.x format
normalize_version() {
    local v="$1"
    v="${v#v}"  # Remove leading 'v' if present

    # Split by dots and filter numeric parts
    IFS='.' read -ra parts <<< "$v"
    local major="${parts[0]:-0}"
    local minor="${parts[1]:-0}"
    local patch="${parts[2]:-0}"

    # Ensure each part is numeric
    [[ "$major" =~ ^[0-9]+$ ]] || major=0
    [[ "$minor" =~ ^[0-9]+$ ]] || minor=0
    [[ "$patch" =~ ^[0-9]+$ ]] || patch=0

    echo "$major.$minor.$patch"
}

# Get current version from main.go
RAW_VERSION=$(grep 'version = "' cmd/inceptor/main.go | sed 's/.*"\(.*\)".*/\1/')
CURRENT_VERSION=$(normalize_version "$RAW_VERSION")
echo "Current version: $CURRENT_VERSION"

# Auto-increment function
increment_version() {
    IFS='.' read -r major minor patch <<< "$CURRENT_VERSION"
    patch=$((patch + 1))
    echo "$major.$minor.$patch"
}

# Set new version
if [ -n "$1" ]; then
    # Check if provided version is valid
    if validate_version "$1"; then
        NEW_VERSION="$1"
    else
        echo "Error: Invalid version format '$1'. Must be x.x.x (e.g., 1.0.1)"
        NEXT_VERSION=$(increment_version)
        echo ""
        read -p "Do you want to release $NEXT_VERSION instead? [y/N] " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            NEW_VERSION="$NEXT_VERSION"
        else
            echo "Aborted."
            exit 1
        fi
    fi
else
    # Auto-increment patch
    NEW_VERSION=$(increment_version)
fi

echo "New version: $NEW_VERSION"

# Update version in main.go
sed -i '' "s/version = \"$CURRENT_VERSION\"/version = \"$NEW_VERSION\"/" cmd/inceptor/main.go

# Build web UI
echo "Building web UI..."
cd web && npm install && npm run generate && cd ..

# Copy static files to embed directory
echo "Copying static files..."
rm -rf internal/api/rest/static/* 2>/dev/null || true
mkdir -p internal/api/rest/static
cp -r web/.output/public/* internal/api/rest/static/

# Build binaries
echo "Building inceptor linux-amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o inceptor-linux-amd64 ./cmd/inceptor

echo "Building inceptor linux-arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags='-s -w' -o inceptor-linux-arm64 ./cmd/inceptor

echo "Building inceptor darwin-arm64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags='-s -w' -o inceptor-darwin-arm64 ./cmd/inceptor

echo "Building inceptor darwin-amd64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags='-s -w' -o inceptor-darwin-amd64 ./cmd/inceptor

# Commit version bump
git add -A
git commit -m "Release v$NEW_VERSION" || true
git push

# Create GitHub release with binaries
echo "Creating GitHub release..."
gh release create "v$NEW_VERSION" \
    inceptor-linux-amd64 \
    inceptor-linux-arm64 \
    inceptor-darwin-arm64 \
    inceptor-darwin-amd64 \
    --title "v$NEW_VERSION" \
    --generate-release-notes

# Cleanup binaries
rm -f inceptor-linux-* inceptor-darwin-*

echo ""
echo "Released v$NEW_VERSION"
echo "GitHub: https://github.com/base-go/inceptor/releases/tag/v$NEW_VERSION"
echo ""
echo "To update your server, run:"
echo "  curl -X POST https://inceptor.common.al/api/v1/system/update -H 'Cookie: session=<token>'"
