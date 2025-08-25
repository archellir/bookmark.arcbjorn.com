#!/bin/bash

# Build script for Torimemo bookmark manager

set -e

echo "🚀 Building Torimemo..."

# Build frontend
echo "📦 Building frontend..."
cd web && pnpm run build && cd ..

# Build backend with optimizations
echo "🔧 Building backend..."
CGO_ENABLED=1 go build -ldflags="-s -w" -o torimemo .

# Get binary size
SIZE=$(du -h torimemo | cut -f1)
echo "✅ Build complete! Binary: ./torimemo (${SIZE})"

# Show next steps
echo ""
echo "🚀 Next steps:"
echo "  • Run locally: ./torimemo"
echo "  • Build Docker: docker build -t torimemo ."
echo "  • Deploy with compose: docker-compose up -d"
echo "  • Deploy to k8s: kubectl apply -f k8s.yaml"