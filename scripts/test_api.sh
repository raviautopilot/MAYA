#!/usr/bin/env bash
# =============================================================================
# MyKanban Backend — API Test Script (curl)
# =============================================================================
# Usage:
#   bash scripts/test_api.sh
#
# Prerequisites:
#   - Server running on localhost:8080 (or set BASE_URL)
#   - curl and jq installed
# =============================================================================

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
EMAIL="${EMAIL:-admin@mykanban.local}"
PASSWORD="${PASSWORD:-admin123}"

# Color helpers
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

header() {
  echo ""
  echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo -e "${YELLOW}  $1${NC}"
  echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

run() {
  echo -e "${GREEN}▶ $1${NC}"
  shift
  "$@" | jq . 2>/dev/null || true
  echo ""
}

# =============================================================================
# 1. HEALTH CHECK
# =============================================================================
header "1. Health Check (no auth required)"

run "GET /api/health" \
  curl -s "${BASE_URL}/api/health"

# =============================================================================
# 2. LOGIN — get JWT token
# =============================================================================
header "2. Login — obtain JWT token"

LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")

echo "$LOGIN_RESPONSE" | jq .

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token // empty')
if [ -z "$TOKEN" ]; then
  echo "ERROR: Failed to get token. Check credentials and server status."
  exit 1
fi
echo -e "${GREEN}✓ Token obtained successfully${NC}"

AUTH="Authorization: Bearer ${TOKEN}"

# =============================================================================
# 3. CHANGE PASSWORD (optional — commented out to avoid locking yourself out)
# =============================================================================
header "3. Change Password (example — commented out)"

echo '# Uncomment to test password change:'
echo '# curl -s -X POST "${BASE_URL}/api/v1/auth/change-password" \'
echo '#   -H "Content-Type: application/json" \'
echo '#   -H "${AUTH}" \'
echo '#   -d '\''{"old_password":"admin123","new_password":"newSecure123"}'\'''

# =============================================================================
# 4. PROJECTS — Full CRUD
# =============================================================================
header "4a. Create Project"

PROJECT_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/projects" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d '{
    "name": "Test Project",
    "description": "Created by test_api.sh",
    "type": "personal"
  }')
echo "$PROJECT_RESPONSE" | jq .
PROJECT_ID=$(echo "$PROJECT_RESPONSE" | jq -r '.data.id // empty')
echo -e "${GREEN}✓ Project ID: ${PROJECT_ID}${NC}"

# ----
header "4b. List Projects"

run "GET /api/v1/projects" \
  curl -s "${BASE_URL}/api/v1/projects" \
  -H "${AUTH}"

# ----
header "4c. Get Project by ID"

run "GET /api/v1/projects/${PROJECT_ID}" \
  curl -s "${BASE_URL}/api/v1/projects/${PROJECT_ID}" \
  -H "${AUTH}"

# ----
header "4d. Update Project"

run "PUT /api/v1/projects/${PROJECT_ID}" \
  curl -s -X PUT "${BASE_URL}/api/v1/projects/${PROJECT_ID}" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d '{
    "name": "Updated Project",
    "description": "Updated by test_api.sh",
    "type": "professional"
  }'

# =============================================================================
# 5. BOARDS — Full CRUD
# =============================================================================
header "5a. Create Board"

BOARD_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/boards" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d "{
    \"project_id\": \"${PROJECT_ID}\",
    \"name\": \"Sprint Board\",
    \"swimlanes\": [\"To Do\", \"In Progress\", \"In Review\", \"Done\"],
    \"task_types\": [\"Bug\", \"Feature\", \"Chore\", \"Spike\"]
  }")
echo "$BOARD_RESPONSE" | jq .
BOARD_ID=$(echo "$BOARD_RESPONSE" | jq -r '.data.id // empty')
echo -e "${GREEN}✓ Board ID: ${BOARD_ID}${NC}"

# ----
header "5b. List Boards (all)"

run "GET /api/v1/boards" \
  curl -s "${BASE_URL}/api/v1/boards" \
  -H "${AUTH}"

# ----
header "5c. List Boards (filtered by project_id)"

run "GET /api/v1/boards?project_id=${PROJECT_ID}" \
  curl -s "${BASE_URL}/api/v1/boards?project_id=${PROJECT_ID}" \
  -H "${AUTH}"

# ----
header "5d. Get Board by ID"

run "GET /api/v1/boards/${BOARD_ID}" \
  curl -s "${BASE_URL}/api/v1/boards/${BOARD_ID}" \
  -H "${AUTH}"

# ----
header "5e. Update Board"

run "PUT /api/v1/boards/${BOARD_ID}" \
  curl -s -X PUT "${BASE_URL}/api/v1/boards/${BOARD_ID}" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d "{
    \"project_id\": \"${PROJECT_ID}\",
    \"name\": \"Updated Sprint Board\",
    \"swimlanes\": [\"To Do\", \"In Progress\", \"In Review\", \"Done\"],
    \"task_types\": [\"Bug\", \"Feature\", \"Chore\"]
  }"

# =============================================================================
# 6. RESOURCES — Full CRUD
# =============================================================================
header "6a. Create Resource (Global)"

RESOURCE_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/resources" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d '{
    "name": "John Doe",
    "type": "Global",
    "linked_items": []
  }')
echo "$RESOURCE_RESPONSE" | jq .
RESOURCE_ID=$(echo "$RESOURCE_RESPONSE" | jq -r '.data.id // empty')
echo -e "${GREEN}✓ Resource ID: ${RESOURCE_ID}${NC}"

# ----
header "6b. List Resources"

run "GET /api/v1/resources" \
  curl -s "${BASE_URL}/api/v1/resources" \
  -H "${AUTH}"

# ----
header "6c. Get Resource by ID"

run "GET /api/v1/resources/${RESOURCE_ID}" \
  curl -s "${BASE_URL}/api/v1/resources/${RESOURCE_ID}" \
  -H "${AUTH}"

# ----
header "6d. Update Resource (link to project)"

run "PUT /api/v1/resources/${RESOURCE_ID}" \
  curl -s -X PUT "${BASE_URL}/api/v1/resources/${RESOURCE_ID}" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d "{
    \"name\": \"John Doe\",
    \"type\": \"Project\",
    \"linked_items\": [\"${PROJECT_ID}\"]
  }"

# =============================================================================
# 7. TASKS — Full CRUD + PATCH
# =============================================================================
header "7a. Create Task"

TASK_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/tasks" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d "{
    \"board_id\": \"${BOARD_ID}\",
    \"swimlane\": \"To Do\",
    \"task_type\": \"Feature\",
    \"title\": \"Implement login page\",
    \"description\": \"Build the login UI with email/password form\",
    \"assignee_id\": \"${RESOURCE_ID}\",
    \"estimation_minutes\": 120,
    \"cost\": 50.0,
    \"priority\": \"High\",
    \"reminders\": [
      {\"time\": \"2026-05-15T09:00:00Z\", \"note\": \"Start working on this\"},
      {\"time\": \"2026-05-16T09:00:00Z\", \"note\": \"Follow up\"}
    ]
  }")
echo "$TASK_RESPONSE" | jq .
TASK_ID=$(echo "$TASK_RESPONSE" | jq -r '.data.id // empty')
echo -e "${GREEN}✓ Task ID: ${TASK_ID}${NC}"

# ----
header "7b. List Tasks (all)"

run "GET /api/v1/tasks" \
  curl -s "${BASE_URL}/api/v1/tasks" \
  -H "${AUTH}"

# ----
header "7c. List Tasks (filtered by board_id)"

run "GET /api/v1/tasks?board_id=${BOARD_ID}" \
  curl -s "${BASE_URL}/api/v1/tasks?board_id=${BOARD_ID}" \
  -H "${AUTH}"

# ----
header "7d. List Tasks (filtered by priority)"

run "GET /api/v1/tasks?priority=High" \
  curl -s "${BASE_URL}/api/v1/tasks?priority=High" \
  -H "${AUTH}"

# ----
header "7e. Get Task by ID"

run "GET /api/v1/tasks/${TASK_ID}" \
  curl -s "${BASE_URL}/api/v1/tasks/${TASK_ID}" \
  -H "${AUTH}"

# ----
header "7f. Patch Task — move to In Progress"

run "PATCH /api/v1/tasks/${TASK_ID} (move swimlane)" \
  curl -s -X PATCH "${BASE_URL}/api/v1/tasks/${TASK_ID}" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d '{"swimlane": "In Progress"}'

# ----
header "7g. Patch Task — update priority and actual time"

run "PATCH /api/v1/tasks/${TASK_ID} (update fields)" \
  curl -s -X PATCH "${BASE_URL}/api/v1/tasks/${TASK_ID}" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d '{"priority": "Critical", "actual_time_minutes": 90}'

# ----
header "7h. Full Update Task (PUT)"

run "PUT /api/v1/tasks/${TASK_ID}" \
  curl -s -X PUT "${BASE_URL}/api/v1/tasks/${TASK_ID}" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d "{
    \"board_id\": \"${BOARD_ID}\",
    \"swimlane\": \"In Review\",
    \"task_type\": \"Feature\",
    \"title\": \"Implement login page (updated)\",
    \"description\": \"Completed login UI\",
    \"assignee_id\": \"${RESOURCE_ID}\",
    \"estimation_minutes\": 120,
    \"actual_time_minutes\": 110,
    \"cost\": 55.0,
    \"priority\": \"High\",
    \"reminders\": []
  }"

# =============================================================================
# 8. SCHEDULERS — Full CRUD
# =============================================================================
header "8a. Create Scheduler (weekly)"

SCHED_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/schedulers" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d '{
    "name": "Weekly Review",
    "type": "weekly"
  }')
echo "$SCHED_RESPONSE" | jq .
SCHED_ID=$(echo "$SCHED_RESPONSE" | jq -r '.data.id // empty')
echo -e "${GREEN}✓ Scheduler ID: ${SCHED_ID}${NC}"

# ----
header "8b. Create Scheduler (cron)"

SCHED_CRON_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/schedulers" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d '{
    "name": "Every Monday 9AM",
    "type": "cron",
    "cron_expression": "0 9 * * 1"
  }')
echo "$SCHED_CRON_RESPONSE" | jq .

# ----
header "8c. List Schedulers"

run "GET /api/v1/schedulers" \
  curl -s "${BASE_URL}/api/v1/schedulers" \
  -H "${AUTH}"

# ----
header "8d. Get Scheduler by ID"

run "GET /api/v1/schedulers/${SCHED_ID}" \
  curl -s "${BASE_URL}/api/v1/schedulers/${SCHED_ID}" \
  -H "${AUTH}"

# ----
header "8e. Update Scheduler"

run "PUT /api/v1/schedulers/${SCHED_ID}" \
  curl -s -X PUT "${BASE_URL}/api/v1/schedulers/${SCHED_ID}" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d '{
    "name": "Bi-Weekly Review",
    "type": "weekly"
  }'

# =============================================================================
# 9. RECURRING TASK — Link scheduler to task and complete it
# =============================================================================
header "9a. Create Task with Scheduler (recurring)"

RECURRING_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/tasks" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d "{
    \"board_id\": \"${BOARD_ID}\",
    \"swimlane\": \"To Do\",
    \"task_type\": \"Chore\",
    \"title\": \"Weekly standup notes\",
    \"description\": \"Document standup items\",
    \"priority\": \"Medium\",
    \"scheduler_id\": \"${SCHED_ID}\"
  }")
echo "$RECURRING_RESPONSE" | jq .
RECURRING_TASK_ID=$(echo "$RECURRING_RESPONSE" | jq -r '.data.id // empty')

# ----
header "9b. Complete Recurring Task — triggers new task generation"

run "PATCH /api/v1/tasks/${RECURRING_TASK_ID} (move to Done)" \
  curl -s -X PATCH "${BASE_URL}/api/v1/tasks/${RECURRING_TASK_ID}" \
  -H "Content-Type: application/json" \
  -H "${AUTH}" \
  -d '{"swimlane": "Done"}'

# ----
header "9c. Verify — new recurring task was created"

run "GET /api/v1/tasks?board_id=${BOARD_ID}&swimlane=To Do" \
  curl -s "${BASE_URL}/api/v1/tasks?board_id=${BOARD_ID}&swimlane=To%20Do" \
  -H "${AUTH}"

# =============================================================================
# 10. SOFT DELETE — Delete entities (soft-delete sets deleted_at)
# =============================================================================
header "10a. Delete Task (soft-delete)"

run "DELETE /api/v1/tasks/${TASK_ID}" \
  curl -s -X DELETE "${BASE_URL}/api/v1/tasks/${TASK_ID}" \
  -H "${AUTH}"

# ----
header "10b. Delete Scheduler (soft-delete)"

run "DELETE /api/v1/schedulers/${SCHED_ID}" \
  curl -s -X DELETE "${BASE_URL}/api/v1/schedulers/${SCHED_ID}" \
  -H "${AUTH}"

# ----
header "10c. Delete Resource (soft-delete)"

run "DELETE /api/v1/resources/${RESOURCE_ID}" \
  curl -s -X DELETE "${BASE_URL}/api/v1/resources/${RESOURCE_ID}" \
  -H "${AUTH}"

# ----
header "10d. Delete Board (soft-delete)"

run "DELETE /api/v1/boards/${BOARD_ID}" \
  curl -s -X DELETE "${BASE_URL}/api/v1/boards/${BOARD_ID}" \
  -H "${AUTH}"

# ----
header "10e. Delete Project (soft-delete)"

run "DELETE /api/v1/projects/${PROJECT_ID}" \
  curl -s -X DELETE "${BASE_URL}/api/v1/projects/${PROJECT_ID}" \
  -H "${AUTH}"

# =============================================================================
# 11. VERIFY DELETIONS — Lists should be empty now
# =============================================================================
header "11. Verify all entities are soft-deleted (lists should be empty)"

run "GET /api/v1/projects (should be empty)" \
  curl -s "${BASE_URL}/api/v1/projects" -H "${AUTH}"

run "GET /api/v1/boards (should be empty)" \
  curl -s "${BASE_URL}/api/v1/boards" -H "${AUTH}"

run "GET /api/v1/tasks (should show only the auto-generated recurring task)" \
  curl -s "${BASE_URL}/api/v1/tasks" -H "${AUTH}"

# =============================================================================
header "✅ All API tests completed!"
echo -e "${GREEN}Server: ${BASE_URL}${NC}"
echo -e "${GREEN}Swagger UI: ${BASE_URL}/swagger/index.html${NC}"
