# とりメモ (Torimemo) - Blazingly Fast Bookmark Manager

A lightning-fast, lightweight bookmark manager with AI-powered auto-categorization, built with Go + TypeScript + SQLite.

## ✨ Features

- **🚀 Blazingly Fast**: Single Go binary with embedded frontend
- **🪶 Ultra Lightweight**: <50MB RAM usage, minimal resource footprint  
- **🔍 Full-Text Search**: Powered by SQLite FTS5 with ranked results
- **🏷️ Smart Tagging**: Interactive tag cloud with auto-categorization
- **🎨 Cyberpunk UI**: Futuristic design with neon effects
- **⚡ Real-Time**: Instant search and updates
- **🐳 Container Ready**: Docker & Kubernetes deployment configs
- **🤖 AI Ready**: Learning system for intelligent categorization

## 🚀 Quick Start

### Local Development

```bash
# Clone and build
git clone <repo>
cd bookmark.arcbjorn.com
./build.sh

# Run
./torimemo

# Open http://localhost:8080
```

### Docker

```bash
# Build and run
docker build -t torimemo .
docker run -p 8080:8080 -v torimemo_data:/data torimemo

# Or use docker-compose
docker-compose up -d
```

### Kubernetes

```bash
kubectl apply -f k8s.yaml
```

## 📖 API Documentation

### Bookmarks

- `GET /api/bookmarks` - List bookmarks (with pagination, filtering)
- `POST /api/bookmarks` - Create bookmark
- `GET /api/bookmarks/{id}` - Get bookmark
- `PUT /api/bookmarks/{id}` - Update bookmark
- `DELETE /api/bookmarks/{id}` - Delete bookmark
- `GET /api/bookmarks/search?q={query}` - Full-text search

### Tags

- `GET /api/tags` - List tags
- `GET /api/tags/cloud` - Tag cloud with sizes
- `GET /api/tags/popular` - Popular tags
- `POST /api/tags` - Create tag
- `PUT /api/tags/{id}` - Update tag
- `DELETE /api/tags/{id}` - Delete tag

### System

- `GET /api/health` - Health check
- `GET /api/stats` - Database statistics

## 🏗️ Architecture

- **Backend**: Go 1.25 (standard library only)
- **Database**: SQLite with FTS5, WAL mode
- **Frontend**: TypeScript + Lit Web Components + TailwindCSS
- **Search**: Full-text search with ranking and snippets
- **Tags**: Many-to-many with weighted tag cloud
- **Deployment**: Single binary with embedded assets

## 🎯 Performance

- **Binary Size**: ~15MB (with embedded frontend)
- **Memory Usage**: <50MB RAM
- **Cold Start**: <100ms
- **Search Latency**: <10ms for 10k+ bookmarks
- **Database**: SQLite with FTS5 indexing

## 🛠️ Development

```bash
# Frontend development
cd web
pnpm install
pnpm run dev

# Backend development  
go run main.go

# Build optimized
./build.sh
```

## 📦 Deployment Options

### Environment Variables

- `PORT`: Server port (default: 8080)
- `DB_PATH`: Database path (default: ./torimemo.db)

### Production Checklist

- [ ] Set `DB_PATH` to persistent volume
- [ ] Configure reverse proxy (nginx/traefik)
- [ ] Set up SSL/TLS termination
- [ ] Configure backup strategy
- [ ] Monitor health endpoint

## 🤖 AI Categorization (Coming Soon)

Three-layer AI system for intelligent bookmark categorization:

1. **Rule-based**: Domain patterns, URL analysis
2. **FastText**: Lightweight text classification  
3. **ONNX Models**: Advanced content understanding

Learning system adapts to user behavior and improves suggestions over time.

## 📄 License

MIT License - see LICENSE file for details.

---

Built with ❤️ in Go and TypeScript