# MyKanban User Journeys

## Journey 1: Admin Setup — First-Time Login & Change Password

### Step 1: Login with default credentials
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@mykanban.local",
  "password": "admin123"
}
```
**Response:**
```json
{
  "data": { "token": "eyJhbGciOiJIUzI1NiIs..." },
  "error": "",
  "status": 200
}
```

### Step 2: Change default password
```http
POST /api/v1/auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "admin123",
  "new_password": "MySecureP@ssw0rd!"
}
```
**Response:**
```json
{
  "data": { "message": "password changed successfully" },
  "error": "",
  "status": 200
}
```

---

## Journey 2: Create a Board — Shopping Project with Custom Swimlanes

### Step 1: Create the "Shopping" project
```http
POST /api/v1/projects
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Shopping",
  "description": "Personal shopping tracking",
  "type": "personal"
}
```
**Response:** Returns the project with a generated UUID.

### Step 2: Create "Monthly Groceries" board with custom swimlanes
```http
POST /api/v1/boards
Authorization: Bearer <token>
Content-Type: application/json

{
  "project_id": "<shopping-project-id>",
  "name": "Monthly Groceries",
  "swimlanes": ["Wishlist", "To Buy", "Purchased"],
  "task_types": ["Grocery", "Household", "Electronics"]
}
```
**Response:** Returns the board with custom swimlanes applied.

---

## Journey 3: Task Lifecycle — Buy Laptop

### Step 1: Create "Self" resource
```http
POST /api/v1/resources
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Self",
  "type": "Global",
  "linked_items": []
}
```

### Step 2: Create "Buy Laptop" task
```http
POST /api/v1/tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "board_id": "<monthly-groceries-board-id>",
  "swimlane": "Wishlist",
  "task_type": "Electronics",
  "title": "Buy Laptop",
  "description": "Research and purchase a new development laptop",
  "assignee_id": "<self-resource-id>",
  "cost": 1500.00,
  "priority": "High",
  "estimation_minutes": 480,
  "reminders": [
    { "time": "2026-06-01T09:00:00Z", "note": "Start researching models" },
    { "time": "2026-06-15T09:00:00Z", "note": "Make final decision and purchase" }
  ]
}
```

### Step 3: Move to "To Buy"
```http
PATCH /api/v1/tasks/<task-id>
Authorization: Bearer <token>
Content-Type: application/json

{
  "swimlane": "To Buy"
}
```

### Step 4: Move to "Purchased" (Done)
```http
PATCH /api/v1/tasks/<task-id>
Authorization: Bearer <token>
Content-Type: application/json

{
  "swimlane": "Purchased",
  "actual_time_minutes": 360
}
```

---

## Journey 4: Recurring Scheduling — Birthday Party

### Step 1: Create a yearly scheduler
```http
POST /api/v1/schedulers
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Annual Birthday Party",
  "type": "yearly",
  "cron_expression": ""
}
```
**Response:** Returns scheduler with `next_run` set 1 year from now.

### Step 2: Create a board for events
```http
POST /api/v1/boards
Authorization: Bearer <token>
Content-Type: application/json

{
  "project_id": "<personal-project-id>",
  "name": "Events",
  "swimlanes": ["To Do", "In Progress", "Done"]
}
```

### Step 3: Create "Birthday Party" task linked to the scheduler
```http
POST /api/v1/tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "board_id": "<events-board-id>",
  "swimlane": "To Do",
  "task_type": "Chore",
  "title": "Plan Birthday Party",
  "description": "Organize annual birthday celebration",
  "priority": "Medium",
  "scheduler_id": "<birthday-scheduler-id>"
}
```

### Step 4: Complete the task — auto-creates next occurrence
```http
PATCH /api/v1/tasks/<birthday-task-id>
Authorization: Bearer <token>
Content-Type: application/json

{
  "swimlane": "Done"
}
```
**Result:** The current task moves to "Done". A **new identical task** is automatically created in the "To Do" swimlane with `scheduler_id` preserved. The scheduler's `next_run` is updated to 1 year later.

### Step 5: Verify the new recurring task
```http
GET /api/v1/tasks?board_id=<events-board-id>&swimlane=To%20Do
Authorization: Bearer <token>
```
**Response:** Shows the auto-generated "Plan Birthday Party" task ready for the next cycle.

---

## Journey 5: Resource Management — Plumber for Home Maintenance

### Step 1: Create "Home Maintenance" project
```http
POST /api/v1/projects
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Home Maintenance",
  "description": "Track home repairs and upkeep",
  "type": "personal"
}
```

### Step 2: Create a board and "Fix Sink" task
```http
POST /api/v1/boards
Authorization: Bearer <token>

{ "project_id": "<home-maint-id>", "name": "Repairs", "swimlanes": ["To Do", "In Progress", "Done"] }
```
```http
POST /api/v1/tasks
Authorization: Bearer <token>

{
  "board_id": "<repairs-board-id>",
  "swimlane": "To Do",
  "task_type": "Chore",
  "title": "Fix Sink",
  "description": "Kitchen sink is leaking",
  "priority": "High"
}
```

### Step 3: Create "Plumber" resource linked to the project and task
```http
POST /api/v1/resources
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Plumber",
  "type": "Project",
  "linked_items": ["<home-maint-project-id>", "<fix-sink-task-id>"]
}
```

### Step 4: View the resource
```http
GET /api/v1/resources/<plumber-id>
Authorization: Bearer <token>
```
**Response:**
```json
{
  "data": {
    "id": "<plumber-id>",
    "name": "Plumber",
    "type": "Project",
    "linked_items": ["<home-maint-project-id>", "<fix-sink-task-id>"],
    "created_at": "...",
    "updated_at": "..."
  },
  "error": "",
  "status": 200
}
```
