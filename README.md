# MyKanban Monorepo

**MyKanban** is a full-stack Kanban board application for project, board, and task management. This monorepo contains the Go API server backend (`chitta`), the React/TypeScript/Vite web application frontend (`maya`), and the end-to-end testing suite (`e2etest`).

---

## 📁 Monorepo Structure

```
mykanban/
├── chitta/           # Go REST API Backend (Gin framework)
│   ├── auth/         # JWT generation/validation, Google OAuth
│   ├── handlers/     # API endpoints and request handlers
│   ├── middleware/   # Gin recovery & JWT auth middlewares
│   ├── models/       # Struct schemas and database entities
│   ├── storage/      # Thread-safe JSON file-based database layer
│   ├── scripts/      # Build, run, unit testing scripts
│   ├── design/       # Architecture sequence diagrams
│   └── docs/         # Swagger/OpenAPI spec documents
├── maya/             # React & TypeScript Frontend (Vite)
│   ├── src/          # Source components, stores, hooks, router
│   ├── public/       # Static assets
│   └── package.json  # NPM dependencies and scripts
├── e2etest/          # Go End-to-End Testing Suite (Selenium & Webdriver)
│   ├── api/          # Organized REST API test suites
│   ├── web/          # Organized Selenium UI test suites
│   ├── reports/      # Auto-generated interactive HTML test dashboards
│   └── *.go          # Shared helper and config library files
├── manage.sh         # Integrated automation orchestrator & DevOps script
├── package.json      # Workspace configurations and run shortcuts
└── README.md         # This file
```

---

## 🚀 Quick Start & Environment Controller

The core of the local monorepo development lifecycle is orchestrated by `./manage.sh`, which automates starting, stopping, and running tests.

### Usage

```bash
./manage.sh [command]
```

### Available Commands

* **`start`**: Concurrently starts the Chitta Go API Server (port `8080`) and the Maya Vite Web Server (port `3000`) in the background, writing their process IDs to `.dev_pids`.
* **`stop`**: Gracefully terminates running servers using `SIGTERM` and clears the pid files.
* **`kill`**: Forcefully terminates backend and frontend servers with `SIGKILL` and sweeps lingering ports to prevent conflicts.
* **`restart`**: Performs a stop (or kill) followed by a start sequence.
* **`status`**: Checks process state, memory metrics, and outputs runtime stats.
* **`logs`**: Tail backend and frontend log outputs.
* **`e2e`**: Runs the complete E2E testing lifecycle:
  1. Purges lingering processes from ports `8080` and `3000`.
  2. Compiles and launches the backend and frontend servers in background mode.
  3. Verifies servers are up, healthy, and accepting socket connections.
  4. Runs the Go/Selenium browser automation test suites.
  5. Tears down services and frees ports cleanly.

---

## 🧪 Testing

### 1. Backend Unit Tests
Runs the Go native unit tests for handlers and database storage:
```bash
npm run chitta:test
```

### 2. End-to-End Tests
Executes unit API checks and browser workflow automation (login, mock consent portal, dashboard assertion, projects/tasks state verification):
```bash
npm run test:e2e
```
*Visual screenshots, debug execution logs, and detailed interactive HTML reports will be generated automatically inside the `e2etest/reports/` directory.*

---

## 🔧 Troubleshooting & CORS Settings

If you encounter `CORS` issues connecting the frontend to the backend:

1. **Verify `chitta/config.json`**: Ensure `allowed_origins` matches the frontend server URL (e.g. `http://localhost:3000`).
   ```json
   "allowed_origins": "http://localhost:3000"
   ```
2. **Multiple Origins**: Separate permitted origins using commas without spaces:
   ```json
   "allowed_origins": "http://localhost:3000,http://localhost:3001"
   ```
3. **Restart Services**: Reload backend settings using `./manage.sh restart`.

---

## 📄 License

This monorepo is private and proprietary. All rights reserved.
