#!/bin/bash

echo "Building React console with /console/ base path..."

# Navigate to admin directory
cd ../console

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    bun install
fi

# Cleaning old build files
echo "Cleaning old build files..."
rm -rf dist

# Build the React app
echo "Building React app..."
bun run build

# Copy the built files to the console directory for embedding
echo "Copying built files to console directory..."
cd ../sailor
rm -rf cmd/sailor/console
cp -r admin/dist cmd/sailor/console

echo "Build complete! The console is now ready to be embedded."
echo "Make sure to rebuild the Go application to include the new console files."