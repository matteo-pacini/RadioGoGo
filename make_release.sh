#!/usr/bin/env bash

set -eo pipefail

rm -rf bin 2> /dev/null

if [ -z "$1" ]; then
    echo "Please provide a version number"
    exit 1
fi

TARGETS=(
    # macOS
    "darwin|arm64"
    "darwin|amd64"
    # Linux
    "linux|386"
    "linux|arm|6"    # Pi 1, Zero, Zero W - ARMv6
    "linux|arm|7"    # Pi 2, 3, 4 (32-bit OS) - ARMv7
    "linux|arm64"    # Pi 3, 4, 5 (64-bit OS)
    "linux|amd64"
    # Windows
    "windows|386"
    "windows|amd64"
)

mkdir bin
touch bin/checksums.txt

for target in "${TARGETS[@]}"; do
    IFS='|' read -ra parts <<< "$target"
    GOOS=${parts[0]}
    GOARCH=${parts[1]}
    GOARM=${parts[2]:-""}

    # Build filename suffix (include GOARM version if specified)
    if [ -n "$GOARM" ]; then
        SUFFIX="${GOARCH}v${GOARM}"
    else
        SUFFIX="${GOARCH}"
    fi

    OUTPUT="bin/radiogogo"

    # Output name: adjust for windows
    if [ "$GOOS" == "windows" ]; then
        OUTPUT+=".exe"
    fi

    echo "Building for $GOOS/$SUFFIX"
    if [ -n "$GOARM" ]; then
        CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH GOARM=$GOARM go build -ldflags="-s -w -X github.com/zi0p4tch0/radiogogo/data.Version=$1" -o $OUTPUT
    else
        CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w -X github.com/zi0p4tch0/radiogogo/data.Version=$1" -o $OUTPUT
    fi

    zip -j "bin/radiogogo_$1_${GOOS}_${SUFFIX}.zip" $OUTPUT
    rm -rf $OUTPUT

    cd bin && shasum -a 256 "radiogogo_$1_${GOOS}_${SUFFIX}.zip" >> "checksums.txt" && cd ..

done
