# Groupie Tracker Filters

A Go + Angular application that fetches artist data from an external API and lets users filter, search, sort and manage artists.

---

## Features

- Display artists with image, name, first album, members, locations
- **Filters**
  - Creation date — range slider
  - First album year — range input
  - Number of members — checkboxes
  - Concert locations — checkboxes (hierarchical match: `washington, usa` matches `seattle, washington, usa`)
- Search by name, sort, pagination
- CRUD operations (Create, Read, Update, Delete)
- Artist detail page


---

## How to Run

### Option 1 — `go run` (local, no Docker)

**Store: in-memory — data is lost on restart.**

```
terminal 1                        terminal 2
──────────────────────────────    ──────────────────────────
go run ./cmd                      cd frontend
                                  npm install
                                  npm start
```

- Backend: http://localhost:8080
- Frontend: http://localhost:4200 (proxies `/api/*` to `:8080`)

> On first startup the server blocks until it preloads data from the external API.

---

### Option 2 — `go run` + PostgreSQL (local, persistent data)

Start only the database via Docker, run the rest locally:

```bash
# terminal 1 — database only
docker-compose up postgres

# terminal 2 — backend with postgres
DB_URL="postgres://groupie:groupie@localhost:5432/groupie?sslmode=disable" go run ./cmd

# terminal 3 — frontend
cd frontend && npm start
# open http://localhost:4200
```

**Store: PostgreSQL — data survives restarts.**

---

### Option 3 — Docker Compose (3 separate services)

```bash
docker-compose up --build
```

**Store: PostgreSQL — data survives restarts.**

| Service    | Role                              | Port (external) |
|------------|-----------------------------------|-----------------|
| `postgres` | PostgreSQL database               | none            |
| `backend`  | Go API                            | none            |
| `frontend` | nginx — serves UI, proxies `/api` | **8080**        |

Open: **http://localhost:8080**

```bash
# run in background
docker-compose up --build -d

# logs
docker-compose logs -f
docker-compose logs -f backend

# stop (data kept)
docker-compose down

# stop + delete database
docker-compose down -v
```

---

### Option 4 — Single Docker image (no PostgreSQL)

Builds everything into one image. Uses in-memory store.

```bash
# build
docker build -t groupie-tracker:latest .

# run
docker run -d -p 8080:8080 --name groupie-tracker groupie-tracker:latest

# open http://localhost:8080

# logs
docker logs -f groupie-tracker

# stop
docker stop groupie-tracker && docker rm groupie-tracker
```

**Store: in-memory — data is lost on restart.**

---

## Comparison

| Mode | Command | Store | Data persists | Ports |
|---|---|---|---|---|
| Local dev | `go run ./cmd` + `npm start` | in-memory | No | 8080 + 4200 |
| Local + DB | `go run` + `docker-compose up postgres` | PostgreSQL | Yes | 8080 + 4200 |
| Docker Compose | `docker-compose up --build` | PostgreSQL | Yes | 8080 |
| Single image | `docker build` + `docker run` | in-memory | No | 8080 |

---

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/api/artists` | List all artists |
| GET | `/api/artists/:id` | Get artist by ID |
| POST | `/api/artists` | Create artist |
| PUT | `/api/artists/:id` | Update artist |
| DELETE | `/api/artists/:id` | Delete artist |
| POST | `/api/artists/filter` | Filter artists |
| POST | `/api/search` | Search by name (async) |
| GET | `/api/loaded` | Check if API data is loaded |
| GET | `/health` | Health check |

---

## Running Tests

```bash
go test ./...
```

---

## Project Structure

```
groupie-tracker-filters/
├── cmd/main.go               # HTTP server entrypoint (:8080)
├── backend/
│   ├── api/                  # Fetches data from external API on startup
│   ├── events/               # Async search worker (goroutines + channels)
│   ├── handlers/             # HTTP handlers + middleware
│   ├── models/               # Artist, FilterRequest structs
│   └── services/             # Business logic layer
├── db/
│   ├── store.go              # Store interface
│   ├── memory.go             # In-memory implementation (default)
│   └── postgres.go           # PostgreSQL implementation (when DB_URL is set)
├── frontend/                 # Angular 21 standalone app
│   ├── nginx.conf            # nginx config (used in Docker)
│   └── proxy.conf.json       # Dev proxy → backend :8080
├── Dockerfile                # Single-image build (frontend + backend, in-memory)
├── Dockerfile.backend        # Backend-only image (used by docker-compose)
├── Dockerfile.frontend       # Frontend nginx image (used by docker-compose)
└── docker-compose.yml        # 3 services: postgres, backend, frontend
```
