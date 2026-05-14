# MyKanban Frontend

A Next.js 14 frontend for the MyKanban personal & professional task tracker.

## Tech Stack

- **Framework:** Next.js 14 (App Router, TypeScript)
- **Styling:** Tailwind CSS
- **State:** Zustand
- **Forms:** React Hook Form
- **HTTP Client:** Axios
- **Icons:** Lucide React

## Getting Started

### Prerequisites

- Node.js 18+
- Backend running at `http://localhost:8080`

### Install & Run

```bash
cd frontend
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

### Default Credentials

- **Email:** `admin@mykanban.local`
- **Password:** `admin123`

## Environment Variables

Create `.env.local`:

```
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

## Project Structure

```
src/
├── app/
│   ├── login/              # Login page
│   ├── (dashboard)/        # Protected layout group
│   │   ├── projects/       # Projects management
│   │   ├── boards/         # Boards management
│   │   ├── tasks/          # Tasks (Kanban + Table)
│   │   ├── schedulers/     # Schedulers management
│   │   └── resources/      # Resources management
│   └── layout.tsx          # Root layout
├── components/
│   ├── ui/                 # Reusable UI components
│   ├── Sidebar.tsx         # Navigation sidebar
│   └── Header.tsx          # Top header bar
├── services/api/           # API client per entity
├── store/                  # Zustand stores
└── types/                  # TypeScript interfaces
```

## Features

- JWT authentication with auto-logout on token expiry
- Google OAuth support
- Kanban board view with drag-to-swimlane
- Full CRUD for Projects, Boards, Tasks, Schedulers, Resources
- Task reminders (max 5)
- Filtering by board, project, priority
- Responsive sidebar navigation
- Toast notifications
- All interactive elements have `e2e-test-id` attributes

## Scripts

```bash
npm run dev      # Development server
npm run build    # Production build
npm run start    # Production server
npm run lint     # ESLint
```
