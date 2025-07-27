#!/bin/bash

# Build and Deploy ADK-Powered Widget Builder
# This script builds both the Go backend and Java ADK service

set -e

echo "🚀 Building ADK-Powered Widget Builder..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Build Java ADK Service
echo "📦 Building Java ADK Service..."
cd adk_service_java
mvn clean package -DskipTests
cd ..

# Build Go Backend
echo "🏗️ Building Go Backend..."
docker build -t homeboard-enhanced:latest .

# Build ADK Service Docker Image
echo "☕ Building ADK Service Docker Image..."
docker build -t homeboard-adk:latest ./adk_service_java

# Stop existing containers
echo "🛑 Stopping existing containers..."
docker-compose down || true

# Start the services
echo "🚀 Starting services..."
docker-compose up -d

echo "✅ Build complete!"
echo ""
echo "Services running:"
echo "  📊 Homeboard Backend: http://localhost:8080"
echo "  🤖 ADK Service: http://localhost:8081"
echo "  🌐 Admin Panel: http://localhost:8080/admin"
echo ""
echo "To view logs: docker-compose logs -f"
echo "To stop: docker-compose down"