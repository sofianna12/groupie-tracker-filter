# Groupie Tracker (Go backend + Angular frontend)

Short guide to run the project locally, how the pieces fit, and troubleshooting notes.

## What it does

- Displays a list of artists (name, image, album, locations)
- Search, sort, pagination
- CRUD operations (Create, Read, Update, Delete)
- Dark theme UI

Repo layout (relevant parts)
- `cmd/` - Go server entrypoint (`cmd/main.go`). Runs HTTP server on `:8080`.
- `internal/backend` - Backend handlers, services and API fetcher.
- `internal/db` - In-memory store used by the server.
- `internal/frontend` - Angular app (standalone components) used as the UI.

## Production Mode (Single Port :8080)

The Go backend serves both the API and the Angular static files on port 8080.

### Steps:

1) Build the Angular frontend:

```bash
cd internal/frontend
npm install
npm run build
```

2) Start the Go server:

```bash
cd groupie-tracker
go run ./cmd
```

3) Open browser at http://localhost:8080

The Go server will:
- Serve API endpoints at `/api/*`
- Serve Angular static files for all other routes
- Handle Angular routing (SPA fallback to index.html)

## Development Mode (Two Ports)

For faster development with hot-reload:

Requirements
- Go 1.25+ installed and available as `go` in PATH.
- Node.js (recommended Node 18+) and npm. Angular CLI optional (we use `npx @angular/cli` when needed).

1) Start the Go backend

```bash
# from repository root
cd /home/pc/zon01/go_files/groupie-tracker
# start the server (binds to :8080)
go run ./cmd
```

- The server logs a startup message: " Server running on http://localhost:8080" when listening.
- On first startup the server will preload external API data synchronously (blocks until data is loaded).

2) Start the frontend (Angular dev server with proxy)

```bash
cd frontend
npm install
npm run start
```

- This uses `ng serve --proxy-config proxy.conf.json`. The dev server runs on http://localhost:4200 by default.
- The `proxy.conf.json` in `internal/frontend` proxies requests starting with `/api` to the Go backend at `http://localhost:8080`, so you can use relative URLs like `/api/artists` in the frontend code and avoid CORS issues during development.

How the frontend talks to backend
- The Angular service `internal/frontend/src/app/services/artist.service.ts` calls endpoints such as:
	- `GET /api/artists` — list all artists
	- `GET /api/artists/:id` — artist details
	- `POST /api/search` — search by name (the backend uses an async search worker)
- Use the dev server proxy when developing (`npm run start`) so `/api` requests are forwarded to the Go server.

Useful commands
- Build frontend for production and run on single port:

```bash
cd frontend
npm run build
cd groupie-tracker
go run ./cmd
# Open http://localhost:8080
```

## Docker Deployment

The easiest way to run the application is using Docker.

### Prerequisites
- Docker installed
- Docker Compose installed (optional but recommended)

### Quick Start with Docker Compose

```bash
# Build and run
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

The application will be available at http://localhost:8080

### Manual Docker Commands

```bash
# Build the image
docker build -t groupie-tracker:latest .

# Run the container
docker run -d -p 8080:8080 --name groupie-tracker groupie-tracker:latest

# View logs
docker logs -f groupie-tracker

# Stop and remove
docker stop groupie-tracker
docker rm groupie-tracker
```

### Using Makefile

```bash
# Build Docker image
make docker-build

# Run with docker-compose
make docker-run

# View logs
make docker-logs

# Stop
make docker-stop

# Clean everything
make docker-clean
```

### Docker Image Details

The Docker image uses multi-stage build:
1. **Stage 1**: Builds Angular frontend (Node.js)
2. **Stage 2**: Builds Go backend
3. **Stage 3**: Creates minimal runtime image (Alpine Linux)

Final image size: ~20-30 MB

### Health Check

The container includes a health check that pings `/health` endpoint every 30 seconds.

Check container health:
```bash
docker ps
# Look for "healthy" status
```
