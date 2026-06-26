#!/bin/bash

# Pritunl CLI Build Script
# Supports building for Linux and macOS platforms
# Usage: ./build.sh [target]
# Targets: all (default), linux, darwin

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="pritunl"
PROJECT_VERSION=$(git describe --tags --always 2>/dev/null || echo "1.0.0")
BUILD_DIR="./dist"
GOFLAGS="-trimpath"

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to build for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local output_name="${PROJECT_NAME}-${os}-${arch}"

    print_info "Building for ${os}/${arch}..."

    GOOS=$os GOARCH=$arch go build $GOFLAGS \
        -o "${BUILD_DIR}/${output_name}" \
        -ldflags "-X main.Version=${PROJECT_VERSION}" \
        main.go

    if [ -f "${BUILD_DIR}/${output_name}" ]; then
        local size=$(du -h "${BUILD_DIR}/${output_name}" | cut -f1)
        print_success "Built ${output_name} (${size})"
        return 0
    else
        print_error "Failed to build ${output_name}"
        return 1
    fi
}

# Function to build all targets
build_all() {
    print_info "Building $PROJECT_NAME v$PROJECT_VERSION"
    print_info "Output directory: $BUILD_DIR"
    echo ""

    # Create build directory
    mkdir -p "$BUILD_DIR"

    # Build for Linux
    print_info "=== Linux Builds ==="
    build_platform "linux" "amd64"
    build_platform "linux" "arm64"
    echo ""

    # Build for macOS
    print_info "=== macOS Builds ==="
    build_platform "darwin" "amd64"
    build_platform "darwin" "arm64"
    echo ""

    # List all built binaries
    print_success "Build complete! Binaries in $BUILD_DIR:"
    ls -lh "$BUILD_DIR"
}

# Function to build Linux only
build_linux() {
    print_info "Building for Linux..."
    mkdir -p "$BUILD_DIR"

    print_info "=== Linux Builds ==="
    build_platform "linux" "amd64"
    build_platform "linux" "arm64"
    echo ""

    print_success "Linux builds complete! Binaries in $BUILD_DIR:"
    ls -lh "$BUILD_DIR"/pritunl-linux-*
}

# Function to build macOS only
build_darwin() {
    print_info "Building for macOS..."
    mkdir -p "$BUILD_DIR"

    print_info "=== macOS Builds ==="
    build_platform "darwin" "amd64"
    build_platform "darwin" "arm64"
    echo ""

    print_success "macOS builds complete! Binaries in $BUILD_DIR:"
    ls -lh "$BUILD_DIR"/pritunl-darwin-*
}

# Function to show usage
show_usage() {
    cat << EOF
${BLUE}Pritunl CLI Build Script${NC}

Usage: ./build.sh [target]

Targets:
  all              Build for all platforms (Linux amd64/arm64, macOS amd64/arm64) - default
  linux            Build for Linux only (amd64, arm64)
  darwin           Build for macOS only (amd64, arm64)
  clean            Remove build directory
  help             Show this help message

Environment Variables:
  GOFLAGS          Additional Go build flags (default: -trimpath)
  BUILD_DIR        Output directory (default: ./dist)

Examples:
  ./build.sh                 # Build all platforms
  ./build.sh linux           # Build Linux only
  ./build.sh darwin          # Build macOS only
  ./build.sh clean           # Clean build directory

Output:
  Linux amd64:   pritunl-linux-amd64
  Linux arm64:   pritunl-linux-arm64
  macOS amd64:   pritunl-darwin-amd64
  macOS arm64:   pritunl-darwin-arm64

EOF
}

# Function to clean build directory
clean_build() {
    print_info "Cleaning build directory..."
    rm -rf "$BUILD_DIR"
    print_success "Build directory cleaned"
}

# Main script logic
main() {
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Get Go version
    GO_VERSION=$(go version | awk '{print $3}')
    print_info "Go version: $GO_VERSION"

    # Parse command line arguments
    TARGET="${1:-all}"

    case "$TARGET" in
        all)
            build_all
            ;;
        linux)
            build_linux
            ;;
        darwin)
            build_darwin
            ;;
        macos)
            # Alias for darwin
            build_darwin
            ;;
        clean)
            clean_build
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            print_error "Unknown target: $TARGET"
            echo ""
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
