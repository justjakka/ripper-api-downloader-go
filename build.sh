#!/bin/bash

if [ ! "$PACKAGE_VERSION" ]; then
    echo -ne "\nMissing Package Version.\n"
    exit 1
fi

# Define the target operating systems and architectures with file extensions
TARGETS=(
  "linux/amd64"
  "darwin/arm64"
  "darwin/amd64"
  "windows/amd64"
)
PACKAGE_NAME="downloader"

# Set the output directory
OUTPUT_DIR="binaries"

# Create the output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Build for each target
for target in "${TARGETS[@]}"; do
  # Split the target into GOOS and GOARCH with file extension
  goos_arch="${target%:*}"
  IFS='/' read -r goos goarch <<< "$goos_arch"

  # Set the environment variables
  export GOOS="$goos"
  export GOARCH="$goarch"

  # Build the package
  echo "Building for $GOOS/$GOARCH..."
  file_name="${PACKAGE_NAME}-${PACKAGE_VERSION}_${GOOS}-${GOARCH}"
  output_file="$OUTPUT_DIR/$file_name"

  if [ "$GOOS" = "windows" ]; then
        output_file+='.exe'
        GOOS=$GOOS GOARCH=$GOARCH go build -o "${output_file}"
    else
        GOOS=$GOOS GOARCH=$GOARCH go build -o "${output_file}"
    fi
done