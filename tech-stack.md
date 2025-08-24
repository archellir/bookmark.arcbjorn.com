# Technology Stack

## Overview
Blazingly fast, lightweight bookmark manager with minimal dependencies.

## Technology Stack
- **Backend**: Go 1.25 standard library only (`net/http`, `database/sql`)
- **Frontend**: Lit + TypeScript (embedded)
- **Database**: SQLite with `modernc.org/sqlite` (pure Go)
- **Build**: Embedded frontend with `go:embed`

## Architecture
```
┌─────────────────────────┐
│   Go Binary (~8MB)      │
├─────────────────────────┤
│  Embedded Static Files  │ ← go:embed directive
│  ├── index.html         │
│  ├── app.js (Lit+TS)    │
│  └── app.css            │
├─────────────────────────┤
│   net/http ServeMux     │ ← Standard library router
│   + JSON encoding       │
├─────────────────────────┤
│   Pure Go SQLite        │ ← modernc.org/sqlite
└─────────────────────────┘
```

## Key Advantages
- Zero external dependencies
- Ultra-lightweight binary
- Built-in HTTP/2 support in Go 1.25
- Fast compilation and startup
- Perfect for containers

## File Structure
```
/cmd/server/main.go
/internal/handlers/
/internal/db/
/web/dist/ (embedded)
```

## Benefits
- Single binary deployment
- No separate frontend server needed
- Blazingly fast (no network calls for assets)
- Perfect for K8s (one container, one process)
- Minimal RAM footprint (typically <100MB)
- Easy Docker deployment with persistent volumes