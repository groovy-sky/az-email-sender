#!/bin/bash
# Cross-platform build script for azemailsender-cli

set -e

# Configuration
APP_NAME="azemailsender-cli"
BUILD_DIR="dist"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}"
DATE="${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"

# Build flags
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# Platforms to build for
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64" 
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

echo "Building ${APP_NAME} version ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"
echo ""

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"
    
    output_name="${APP_NAME}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "Building for ${GOOS}/${GOARCH}..."
    
    env GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -trimpath \
        -ldflags "${LDFLAGS}" \
        -o "${BUILD_DIR}/${output_name}" \
        "./cmd/${APP_NAME}"
    
    if [ $? -ne 0 ]; then
        echo "Failed to build for ${GOOS}/${GOARCH}"
        exit 1
    fi
done

echo ""
echo "Build complete! Binaries available in ${BUILD_DIR}/"
ls -la "${BUILD_DIR}/"

# Create checksums
echo ""
echo "Generating checksums..."
cd "${BUILD_DIR}"
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum * > checksums.txt
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 * > checksums.txt
else
    echo "Warning: No checksum tool found (sha256sum or shasum)"
fi

echo "Build script completed successfully!"