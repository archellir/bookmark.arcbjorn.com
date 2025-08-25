#!/bin/bash

# Torimemo Production Deployment Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Torimemo Production Deployment${NC}"
echo "======================================="

# Check if docker and docker-compose are installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    exit 1
fi

# Check if .env.production exists
if [ ! -f .env.production ]; then
    echo -e "${YELLOW}⚠️  .env.production not found, copying from example${NC}"
    cp .env.example .env.production
    echo -e "${RED}❗ Please edit .env.production with your production values before proceeding${NC}"
    exit 1
fi

# Check if SSL certificates exist (for production)
if [ ! -d "ssl" ]; then
    echo -e "${YELLOW}⚠️  SSL directory not found, creating self-signed certificate for testing${NC}"
    mkdir -p ssl
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout ssl/key.pem \
        -out ssl/cert.pem \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
    echo -e "${YELLOW}⚠️  Using self-signed certificate. Replace with real SSL certificates for production${NC}"
fi

# Build frontend
echo -e "${GREEN}🔨 Building frontend...${NC}"
cd web
npm install
npm run build
cd ..

# Stop existing containers
echo -e "${YELLOW}🛑 Stopping existing containers...${NC}"
docker-compose -f docker-compose.prod.yml down || true

# Build and start containers
echo -e "${GREEN}🏗️  Building and starting containers...${NC}"
docker-compose -f docker-compose.prod.yml up --build -d

# Wait for services to be healthy
echo -e "${GREEN}🩺 Waiting for health checks...${NC}"
sleep 10

# Check if services are running
if docker-compose -f docker-compose.prod.yml ps | grep -q "Up"; then
    echo -e "${GREEN}✅ Deployment successful!${NC}"
    echo ""
    echo -e "${GREEN}🌐 Application URLs:${NC}"
    echo "   HTTP:  http://localhost"
    echo "   HTTPS: https://localhost (with self-signed cert)"
    echo "   API:   https://localhost/api/health"
    echo ""
    echo -e "${GREEN}📊 Useful commands:${NC}"
    echo "   View logs:     docker-compose -f docker-compose.prod.yml logs -f"
    echo "   Stop services: docker-compose -f docker-compose.prod.yml down"
    echo "   View status:   docker-compose -f docker-compose.prod.yml ps"
else
    echo -e "${RED}❌ Deployment failed! Check logs:${NC}"
    docker-compose -f docker-compose.prod.yml logs
    exit 1
fi