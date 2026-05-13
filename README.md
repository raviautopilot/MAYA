# MyKanban Backend

A minimalistic, secure REST API for personal and professional task tracking. Built with Go and the Gin framework, using flat JSON files for storage — no external database required.

## Features

- **5 Entity Types:** Projects, Boards, Tasks, Schedulers, Resources
- **Full CRUD** with soft-delete support for all entities
- **Recurring Tasks:** Automatic task generation when completing scheduled tasks
- **JWT Authentication** with bcrypt-hashed passwords
- **Google OAuth 2.0** integration
- **Thread-safe** JSON file storage using `sync.RWMutex`
- **Swagger** API documentation (via Swaggo)
- **Graceful shutdown** on SIGINT/SIGTERM

## Quick Start

### Prerequisites
- Go 1.21+
- (Optional) `jq` for script helpers

### Setup & Run

```bash
# 1. Build the project
bash scripts/build.sh

# 2. Start the server
bash scripts/run.sh
```

The server starts on port `8080` by default (configurable in `config.json`).

### Default Login Credentials

| Field    | Value                  |
|----------|------------------------|
| Email    | `admin@mykanban.local` |
| Password | `admin123`             |

**⚠️ Change the default password immediately after first login!**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@mykanban.local","password":"admin123"}'
```

## Project Structure

```
mykanban-backend/
├── main.go              # Entry point, router setup, graceful shutdown
├── config.json          # Application configuration
├── config.example.json  # Template configuration
├── go.mod               # Go module definition
├── auth/                # JWT generation/validation, Google OAuth
├── middleware/           # JWT auth middleware, panic recovery
├── handlers/            # HTTP route handlers (CRUD for all entities)
├── models/              # Data structures and validation
├── storage/             # Generic JSON file store with mutex locking
├── scripts/             # Build, test, run, stop, troubleshoot scripts
├── docs/                # Architecture diagrams, ERD, user journeys
└── bin/                 # Compiled binary (generated)
```

## API Endpoints

Base path: `/api/v1`

| Method | Endpoint                     | Auth | Description              |
|--------|------------------------------|------|--------------------------|
| GET    | `/api/health`                | No   | Health check             |
| POST   | `/api/v1/auth/login`         | No   | Login, get JWT           |
| GET    | `/api/v1/auth/google/login`  | No   | Google OAuth redirect    |
| GET    | `/api/v1/auth/google/callback`| No  | Google OAuth callback    |
| POST   | `/api/v1/auth/change-password`| Yes | Change root password     |
| CRUD   | `/api/v1/projects[/:id]`     | Yes  | Project management       |
| CRUD   | `/api/v1/boards[/:id]`       | Yes  | Board management         |
| CRUD   | `/api/v1/tasks[/:id]`        | Yes  | Task management          |
| PATCH  | `/api/v1/tasks/:id`          | Yes  | Partial update / move    |
| CRUD   | `/api/v1/schedulers[/:id]`   | Yes  | Scheduler management     |
| CRUD   | `/api/v1/resources[/:id]`    | Yes  | Resource management      |

## Scripts

| Script                    | Purpose                                |
|---------------------------|----------------------------------------|
| `scripts/build.sh`       | Format, vet, swagger gen, compile      |
| `scripts/test.sh`        | Run tests with coverage report         |
| `scripts/run.sh`         | Start server (auto-builds if needed)   |
| `scripts/stop.sh`        | Gracefully stop the server             |
| `scripts/troubleshoot.sh`| Diagnose port, JSON, Go environment    |

## Configuration

Edit `config.json`:

```json
{
  "port": 8080,
  "root_email": "admin@mykanban.local",
  "root_password_hash": "<bcrypt hash>",
  "jwt_secret": "your-secret-here",
  "jwt_expiry_hours": 24,
  "google_client_id": "",
  "google_client_secret": "",
  "google_redirect_url": "http://localhost:8080/api/v1/auth/google/callback",
  "storage_dir": "./storage",
  "log_file": "server.log"
}
```

## Documentation

- **Architecture:** `docs/architecture.puml`
- **Data Model ERD:** `docs/data_model.puml`
- **User Journeys:** `docs/user_journeys.md`

## License

MIT
