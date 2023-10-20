#!/bin/bash

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
    "linux|arm"
    "linux|arm64"
    "linux|amd64"
    # Windows
    "windows|386"
    "windows|amd64"
    "windows|arm"
    "windows|arm64"
    # FreeBSD
    "freebsd|386"
    "freebsd|arm"
    "freebsd|amd64"
    "freebsd|arm64"
    # OpenBSD
    "openbsd|386"
    "openbsd|arm"
    "openbsd|amd64"
    "openbsd|arm64"
    # NetBSD
    "netbsd|386"
    "netbsd|arm"
    "netbsd|amd64"
    "netbsd|arm64"
)

mkdir bin
touch bin/checksums.txt

for target in "${TARGETS[@]}"; do
    IFS='|' read -ra target <<< "$target"
    GOOS=${target[0]}
    GOARCH=${target[1]}
    
    OUTPUT="bin/radiogogo"

    # Output name: adjust for windows
    if [ "$GOOS" == "windows" ]; then
        OUTPUT+=".exe"
    fi

    echo "Building for $GOOS/$GOARCH"
    GOOS=$GOOS GOARCH=$GOARCH go build -o $OUTPUT

    zip -j "bin/radiogogo_$1_${GOOS}_${GOARCH}.zip" $OUTPUT
    rm -rf $OUTPUT

    cd bin && shasum -a 256 "radiogogo_$1_${GOOS}_${GOARCH}.zip" >> "checksums.txt" && cd ..

done
