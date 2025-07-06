#!/bin/bash

# Set the image name and tag
IMAGE_NAME="sailor"
TAG="latest"

echo "Building Docker image with embedded Sailor Console..."

# Build the Docker image
docker build -t ${IMAGE_NAME}:${TAG} .

if [ $? -eq 0 ]; then
    echo "✅ Docker image built successfully!"
    echo ""
    echo "To run the container:"
    echo "docker run -p 7766:7766 -v \$(pwd)/configs:/app/configs ${IMAGE_NAME}:${TAG}"
    echo ""
    echo "Or run it now? (y/n)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo "Starting container..."
        docker run -p 7766:7766 -v "$(pwd)/configs:/app/configs" ${IMAGE_NAME}:${TAG}
    fi
else
    echo "❌ Docker build failed!"
    exit 1
fi