# MyKanban

**MyKanban** is a full-stack Kanban board application for project and task management. This monorepo contains all components of the application.

---

## 📁 Monorepo Structure

```
mykanban/
├── backend/          # Go REST API server (Gin framework)
│   ├── auth/         # Authentication & JWT handling
│   ├── handlers/     # API route handlers
│   ├── middleware/    # Gin middleware (JWT auth, recovery)
│   ├── models/       # Data models and structs
│   ├── storage/      # JSON file-based storage layer
│   ├── scripts/      # Build, run, test, and utility scripts
│   ├── design/       # Architecture diagrams and design docs
│   ├── docs/         # Swagger/OpenAPI generated docs
│   └── main.go       # Application entry point
├── frontend/         # Frontend application (coming soon)
├── e2e/              # End-to-end tests
├── utils/            # Shared utilities and scripts
├── package.json      # Monorepo workspace configuration
├── CONTRIBUTING.md   # Development workflow guide
└── README.md         # This file
```

## 🚀 Quick Start

### Backend

```bash
cd backend
cp config.example.json config.json   # Configure settings
bash scripts/build.sh                # Build the server
bash scripts/run.sh                  # Start the server
```

The API server runs at `http://localhost:8080` by default. Swagger UI is available at `http://localhost:8080/swagger/index.html`.

### Frontend

> 🚧 Coming soon. See [frontend/README.md](frontend/README.md) for planned details.

### Running Tests

```bash
# Backend unit tests
cd backend
bash scripts/test.sh

# API integration tests
cd backend
bash scripts/test_api.sh
```

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [Backend README](backend/README.md) | Backend setup, API overview, and development guide |
| [Swagger Docs](backend/SWAGGER.md) | How to access and use the Swagger UI |
| [API Tests](backend/api-tests.http) | HTTP client test collection |
| [Architecture](backend/design/architecture.puml) | System architecture diagram |
| [Data Model](backend/design/data_model.puml) | Entity relationship diagram |
| [User Journeys](backend/design/user_journeys.md) | User interaction scenarios |
| [Contributing](CONTRIBUTING.md) | Development workflow for this monorepo |

## 🛠️ Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go, Gin, JWT, bcrypt, Swagger |
| Storage | JSON flat-file with thread-safe access |
| Frontend | *(Planned)* React, TypeScript, Vite |
| Testing | Go test, curl scripts, HTTP client files |

## 📄 License

This project is private. All rights reserved.
