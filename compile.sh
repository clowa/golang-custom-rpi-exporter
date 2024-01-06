#!/bin/bash

# Check if platform arguments are provided
if [ $# -eq 0 ]; then
    echo "Please provide platform(s) as arguments. Example: ./build_all.sh linux/amd64 windows/amd64 darwin/amd64"
    exit 1
fi

# Build for each platform provided as arguments
for platform in "$@"
do
    export GOOS=${platform%/*}
    export GOARCH=${platform#*/}
    output_name="./bin/golang-custom-rpi-exporter-${GOOS}-${GOARCH}"
    if [ $GOOS = "windows" ]; then
        output_name+=".exe"
    fi

    echo "Building for $GOOS/$GOARCH ..."
    go build -o $output_name .
done
