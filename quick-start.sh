#!/bin/bash

# Torimemo Quick Start Script
# Automatically sets up demo data and starts the server

set -e

echo "🚀 とりメモ (Torimemo) Quick Start"
echo "=================================="
echo ""

# Check if binary exists
if [ ! -f "./torimemo" ]; then
    echo "📦 Building Torimemo..."
    ./build.sh
    echo ""
fi

# Check if database exists
if [ ! -f "./torimemo.db" ]; then
    echo "🌱 Setting up demo data..."
    if [ ! -f "./seed" ]; then
        echo "Building seed tool..."
        go build -o seed ./cmd/seed
    fi
    ./seed
    echo ""
fi

# Check if port 8080 is available
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "⚠️  Port 8080 is already in use. Stopping existing process..."
    pkill -f torimemo || true
    sleep 2
fi

echo "🎯 Starting Torimemo server..."
echo ""

# Start the server in background
./torimemo &
SERVER_PID=$!

echo "✅ Server started with PID: $SERVER_PID"
echo ""

# Wait for server to be ready
echo "⏳ Waiting for server to be ready..."
for i in {1..10}; do
    if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
        break
    fi
    sleep 1
done

# Check if server is responding
if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "🎉 Torimemo is ready!"
    echo ""
    echo "📊 Server Info:"
    echo "   • URL: http://localhost:8080"
    echo "   • API: http://localhost:8080/api/health"
    echo "   • Database: ./torimemo.db"
    echo "   • Process ID: $SERVER_PID"
    echo ""
    echo "📱 Features Available:"
    echo "   • 10+ demo bookmarks with tags"
    echo "   • Full-text search (try 'programming')"
    echo "   • Interactive tag cloud"
    echo "   • Advanced search filters"
    echo "   • Export/import functionality"
    echo "   • Analytics dashboard"
    echo ""
    echo "🛑 To stop: pkill -f torimemo"
    echo "📖 Documentation: README.md"
    echo ""
    echo "🌐 Opening browser..."
    
    # Try to open browser (works on most systems)
    if command -v xdg-open > /dev/null; then
        xdg-open http://localhost:8080
    elif command -v open > /dev/null; then
        open http://localhost:8080
    elif command -v start > /dev/null; then
        start http://localhost:8080
    else
        echo "   Please open http://localhost:8080 in your browser"
    fi
    
    echo ""
    echo "✨ Enjoy your blazingly fast bookmark manager!"
    echo ""
    
    # Keep script running to show logs
    echo "📋 Server logs (Ctrl+C to exit):"
    echo "================================"
    wait $SERVER_PID
    
else
    echo "❌ Server failed to start properly"
    echo "   Check the logs above for errors"
    exit 1
fi