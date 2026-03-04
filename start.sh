#!/bin/bash

# ArcaneLink Quick Start Script

echo "🚀 Starting ArcaneLink..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed. Please install docker-compose first."
    exit 1
fi

# Start backend services
echo "📦 Starting backend services..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if services are running
if docker-compose ps | grep -q "Up"; then
    echo "✅ Backend services are running"
else
    echo "❌ Failed to start backend services"
    exit 1
fi

# Start web client
echo "🌐 Starting web client..."
cd web-client
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    npm install
fi

npm run dev &
WEB_PID=$!

echo ""
echo "✅ ArcaneLink is ready!"
echo ""
echo "📍 Services:"
echo "   - API Gateway: http://localhost:8080"
echo "   - Web Client:  http://localhost:3000"
echo ""
echo "Press Ctrl+C to stop all services"

# Trap Ctrl+C
trap "echo ''; echo '🛑 Stopping services...'; kill $WEB_PID; docker-compose down; exit" INT

# Wait
wait $WEB_PID
