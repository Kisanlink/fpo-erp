#!/bin/bash

# Deployment script for Kisanlink ERP

set -e

echo "Deploying Kisanlink ERP..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    exit 1
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "Error: docker-compose is not installed"
    exit 1
fi

# Build Docker image
echo "Building Docker image..."
docker build -t kisanlink-erp:latest .

# Stop existing containers
echo "Stopping existing containers..."
docker-compose down

# Start services
echo "Starting services..."
docker-compose up -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 30

# Check service health
echo "Checking service health..."
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Application is healthy"
else
    echo "❌ Application health check failed"
    exit 1
fi

echo "Deployment completed successfully!"
echo "Application is running at http://localhost:8080"
echo "API documentation at http://localhost:8080/swagger"


