# Contributing to MyKanban

Thank you for contributing to MyKanban! This guide covers the development workflow for this monorepo.

---

## 📁 Repository Layout

```
mykanban/
├── backend/      # Go REST API (Gin)
├── frontend/     # React/TypeScript frontend (coming soon)
├── e2e/          # End-to-end tests
└── utils/        # Shared utilities
```

## 🛠️ Prerequisites

- **Go 1.23+** — for the backend
- **Node.js 18+** and **npm** — for the frontend and monorepo scripts
- **Git** — for version control

## 🚀 Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/raviautopilot/mycan.git
cd mycan
```

### 2. Backend setup

```bash
cd backend
cp config.example.json config.json
# Edit config.json with your settings
bash scripts/build.sh
bash scripts/run.sh
```

### 3. Frontend setup (when available)

```bash
cd frontend
npm install
npm run dev
```

## 🔀 Development Workflow

### Branching

- `main` — stable release branch
- `feature/<name>` — for new features
- `fix/<name>` — for bug fixes
- `refactor/<name>` — for refactoring

### Making Changes

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/my-feature main
   ```

2. Make changes in the appropriate directory:
   - Backend changes → `backend/`
   - Frontend changes → `frontend/`
   - Shared utilities → `utils/`

3. Run relevant tests before committing:
   ```bash
   # Backend tests
   cd backend && bash scripts/test.sh

   # Frontend tests (when available)
   cd frontend && npm test
   ```

4. Commit using conventional commit messages:
   ```
   feat: add new board endpoint
   fix: resolve task sorting bug
   docs: update API documentation
   refactor: restructure handler code
   test: add board CRUD tests
   ```

5. Push and open a Pull Request against `main`.

### Code Review

- All PRs require at least one review
- Ensure CI checks pass (tests, linting, build)
- Keep PRs focused — one feature or fix per PR

## 🧪 Testing

### Backend

```bash
cd backend

# Unit tests with coverage
bash scripts/test.sh

# API integration tests (requires running server)
bash scripts/test_api.sh

# Troubleshoot issues
bash scripts/troubleshoot.sh
```

### E2E Tests

> Coming soon. See [e2e/README.md](e2e/README.md).

## 📝 Documentation

- Update `backend/README.md` for backend-specific changes
- Update `frontend/README.md` for frontend-specific changes
- Update root `README.md` for structural or cross-cutting changes
- Keep Swagger annotations up to date when modifying API endpoints
- Regenerate Swagger docs: `cd backend && swag init --parseDependency --parseInternal -g main.go`

## ❓ Questions?

Open an issue on [GitHub](https://github.com/raviautopilot/mycan/issues).
