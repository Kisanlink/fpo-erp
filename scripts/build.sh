#!/bin/bash

# Build script for Kisanlink ERP

set -e

echo "Building Kisanlink ERP..."

# Set environment
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

# Build the application
echo "Compiling application..."
go build -o bin/kisanlink-erp cmd/server/main.go

echo "Build completed successfully!"
echo "Binary location: bin/kisanlink-erp"

# Optional: Create a tar.gz for distribution
if [ "$1" = "--package" ]; then
    echo "Creating distribution package..."
    mkdir -p dist
    tar -czf dist/kisanlink-erp.tar.gz bin/kisanlink-erp
    echo "Package created: dist/kisanlink-erp.tar.gz"
fi


