package handlers

import (
	"chitta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// ---- Task Handlers ----

// CreateTask creates a new task on a board.
// @Summary      Create task
// @Description  Create a new task. Validates board_id, swimlane, and task_type against the board. Max 5 reminders.
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.Task true "Task data (board_id, swimlane, task_type, title, priority required)"
// @Success      201 {object} models.APIResponse{data=models.Task} "Created task"
// @Failure      400 {object} models.APIResponse "Invalid request, invalid board/swimlane/task_type, or too many reminders"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/tasks [post]
func (h *Handler) CreateTask(c *gin.Context) {
	var t models.Task
	if err := c.ShouldBindJSON(&t); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}
	if len(t.Reminders) > models.MaxReminders {
		respond(c, http.StatusBadRequest, nil, fmt.Sprintf("maximum %d reminders allowed", models.MaxReminders))
		return
	}

	// Validate board exists and swimlane/task_type are valid
	boards, err := h.Boards.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to verify board: "+err.Error())
		return
	}
	var board *models.Board
	for _, b := range boards {
		if b.ID == t.BoardID && b.DeletedAt == nil {
			board = &b
			break
		}
	}
	if board == nil {
		respond(c, http.StatusBadRequest, nil, "board_id does not reference a valid board")
		return
	}
	if !contains(board.Swimlanes, t.Swimlane) {
		respond(c, http.StatusBadRequest, nil, fmt.Sprintf("swimlane '%s' is not valid for this board; valid: %v", t.Swimlane, board.Swimlanes))
		return
	}
	if !contains(board.TaskTypes, t.TaskType) {
		respond(c, http.StatusBadRequest, nil, fmt.Sprintf("task_type '%s' is not valid for this board; valid: %v", t.TaskType, board.TaskTypes))
		return
	}

	t.ID = uuid.New().String()
	t.CreatedAt = time.Now().UTC()
	t.UpdatedAt = t.CreatedAt
	if t.Reminders == nil {
		t.Reminders = []models.Reminder{}
	}

	// Initialize DueDate from Scheduler if linked and not explicitly provided
	if t.SchedulerID != "" && t.DueDate == "" {
		schedulers, err := h.Schedulers.LoadAll()
		if err == nil {
			for _, s := range schedulers {
				if s.ID == t.SchedulerID && s.DeletedAt == nil {
					t.DueDate = s.NextRun
					break
				}
			}
		}
	}

	if err := h.Tasks.WithLock(func(items []models.Task) ([]models.Task, error) {
		return append(items, t), nil
	}); err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to save task: "+err.Error())
		return
	}
	respond(c, http.StatusCreated, t, "")
}

// ListTasks returns all active tasks with optional filters.
// @Summary      List tasks
// @Description  Get all active tasks. Filter by board_id, swimlane, assignee_id, or priority.
// @Tags         Tasks
// @Produce      json
// @Security     BearerAuth
// @Param        board_id    query string false "Filter by board UUID"
// @Param        swimlane    query string false "Filter by swimlane name"
// @Param        assignee_id query string false "Filter by assignee resource UUID"
// @Param        priority    query string false "Filter by priority (Low, Medium, High, Critical)"
// @Success      200 {object} models.APIResponse{data=[]models.Task} "List of tasks"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/tasks [get]
func (h *Handler) ListTasks(c *gin.Context) {
	items, err := h.Tasks.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to load tasks: "+err.Error())
		return
	}
	active := filterActive(items, func(t models.Task) *time.Time { return t.DeletedAt })

	// Apply query filters
	if bid := c.Query("board_id"); bid != "" {
		var filtered []models.Task
		for _, t := range active {
			if t.BoardID == bid {
				filtered = append(filtered, t)
			}
		}
		active = filtered
	}
	if sw := c.Query("swimlane"); sw != "" {
		var filtered []models.Task
		for _, t := range active {
			if t.Swimlane == sw {
				filtered = append(filtered, t)
			}
		}
		active = filtered
	}
	if aid := c.Query("assignee_id"); aid != "" {
		var filtered []models.Task
		for _, t := range active {
			if t.AssigneeID == aid {
				filtered = append(filtered, t)
			}
		}
		active = filtered
	}
	if p := c.Query("priority"); p != "" {
		var filtered []models.Task
		for _, t := range active {
			if t.Priority == p {
				filtered = append(filtered, t)
			}
		}
		active = filtered
	}
	if active == nil {
		active = []models.Task{}
	}
	respond(c, http.StatusOK, active, "")
}

// GetTask retrieves a single task by ID.
// @Summary      Get task
// @Description  Get a task by its UUID
// @Tags         Tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task UUID"
// @Success      200 {object} models.APIResponse{data=models.Task} "Task details"
// @Failure      404 {object} models.APIResponse "Task not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/tasks/{id} [get]
func (h *Handler) GetTask(c *gin.Context) {
	id := c.Param("id")
	items, err := h.Tasks.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to load tasks: "+err.Error())
		return
	}
	for _, t := range items {
		if t.ID == id && t.DeletedAt == nil {
			respond(c, http.StatusOK, t, "")
			return
		}
	}
	respond(c, http.StatusNotFound, nil, "task not found")
}

// UpdateTask replaces a task's data (full update).
// @Summary      Update task
// @Description  Full update of a task by UUID. Max 5 reminders.
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string      true "Task UUID"
// @Param        body body models.Task  true "Updated task data"
// @Success      200 {object} models.APIResponse{data=models.Task} "Updated task"
// @Failure      400 {object} models.APIResponse "Invalid request or too many reminders"
// @Failure      404 {object} models.APIResponse "Task not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/tasks/{id} [put]
func (h *Handler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}
	if len(input.Reminders) > models.MaxReminders {
		respond(c, http.StatusBadRequest, nil, fmt.Sprintf("maximum %d reminders allowed", models.MaxReminders))
		return
	}

	var updated *models.Task
	err := h.Tasks.WithLock(func(items []models.Task) ([]models.Task, error) {
		for i, t := range items {
			if t.ID == id && t.DeletedAt == nil {
				items[i].Title = input.Title
				items[i].Description = input.Description
				items[i].BoardID = input.BoardID
				items[i].Swimlane = input.Swimlane
				items[i].TaskType = input.TaskType
				items[i].AssigneeID = input.AssigneeID
				items[i].EstimationMinutes = input.EstimationMinutes
				items[i].ActualTimeMinutes = input.ActualTimeMinutes
				items[i].Cost = input.Cost
				items[i].Priority = input.Priority
				items[i].SchedulerID = input.SchedulerID
				items[i].DueDate = input.DueDate
				if input.Reminders != nil {
					items[i].Reminders = input.Reminders
				}
				items[i].UpdatedAt = time.Now().UTC()
				updated = &items[i]
				return items, nil
			}
		}
		return items, fmt.Errorf("not found")
	})
	if err != nil {
		if err.Error() == "not found" {
			respond(c, http.StatusNotFound, nil, "task not found")
		} else {
			respond(c, http.StatusInternalServerError, nil, "failed to update task: "+err.Error())
		}
		return
	}
	respond(c, http.StatusOK, updated, "")
}

// PatchTask handles partial updates, including moving swimlane and recurring task generation.
// @Summary      Patch task
// @Description  Partial update of a task. Supports updating individual fields (swimlane, title, description, assignee_id, priority, task_type, cost, estimation_minutes, actual_time_minutes, scheduler_id). Moving to "Done" with a scheduler_id triggers recurring task generation.
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string true "Task UUID"
// @Param        body body object true "Fields to update (only include fields you want to change)"
// @Success      200 {object} models.APIResponse{data=models.Task} "Patched task"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      404 {object} models.APIResponse "Task not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/tasks/{id} [patch]
func (h *Handler) PatchTask(c *gin.Context) {
	id := c.Param("id")
	var patch map[string]interface{}
	if err := c.ShouldBindJSON(&patch); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}

	var patched *models.Task
	err := h.Tasks.WithLock(func(items []models.Task) ([]models.Task, error) {
		for i, t := range items {
			if t.ID == id && t.DeletedAt == nil {
				if v, ok := patch["swimlane"].(string); ok {
					items[i].Swimlane = v
				}
				if v, ok := patch["title"].(string); ok {
					items[i].Title = v
				}
				if v, ok := patch["description"].(string); ok {
					items[i].Description = v
				}
				if v, ok := patch["assignee_id"].(string); ok {
					items[i].AssigneeID = v
				}
				if v, ok := patch["priority"].(string); ok {
					items[i].Priority = v
				}
				if v, ok := patch["task_type"].(string); ok {
					items[i].TaskType = v
				}
				if v, ok := patch["cost"].(float64); ok {
					items[i].Cost = v
				}
				if v, ok := patch["estimation_minutes"].(float64); ok {
					items[i].EstimationMinutes = int(v)
				}
				if v, ok := patch["actual_time_minutes"].(float64); ok {
					items[i].ActualTimeMinutes = int(v)
				}
				if v, ok := patch["scheduler_id"].(string); ok {
					items[i].SchedulerID = v
				}
				if v, ok := patch["due_date"].(string); ok {
					items[i].DueDate = v
				}
				items[i].UpdatedAt = time.Now().UTC()
				patched = &items[i]

				// Recurring task generation: if moved to "Done" and has a scheduler
				if sw, ok := patch["swimlane"].(string); ok && sw == "Done" && items[i].SchedulerID != "" {
					newTask := h.generateRecurringTask(items[i])
					if newTask != nil {
						items = append(items, *newTask)
					}
				}
				return items, nil
			}
		}
		return items, fmt.Errorf("not found")
	})
	if err != nil {
		if err.Error() == "not found" {
			respond(c, http.StatusNotFound, nil, "task not found")
		} else {
			respond(c, http.StatusInternalServerError, nil, "failed to patch task: "+err.Error())
		}
		return
	}
	respond(c, http.StatusOK, patched, "")
}

// DeleteTask soft-deletes a task.
// @Summary      Delete task
// @Description  Soft-delete a task by UUID
// @Tags         Tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Task UUID"
// @Success      200 {object} models.APIResponse{data=object{message=string}} "Deletion confirmation"
// @Failure      404 {object} models.APIResponse "Task not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/tasks/{id} [delete]
func (h *Handler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	now := time.Now().UTC()
	var deleted bool
	err := h.Tasks.WithLock(func(items []models.Task) ([]models.Task, error) {
		for i, t := range items {
			if t.ID == id && t.DeletedAt == nil {
				items[i].DeletedAt = &now
				items[i].UpdatedAt = now
				deleted = true
				return items, nil
			}
		}
		return items, nil
	})
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to delete task: "+err.Error())
		return
	}
	if !deleted {
		respond(c, http.StatusNotFound, nil, "task not found")
		return
	}
	respond(c, http.StatusOK, gin.H{"message": "task deleted"}, "")
}

// generateRecurringTask creates a new task based on a completed task's scheduler.
func (h *Handler) generateRecurringTask(completed models.Task) *models.Task {
	schedulers, err := h.Schedulers.LoadAll()
	if err != nil {
		return nil
	}
	var sched *models.Scheduler
	for _, s := range schedulers {
		if s.ID == completed.SchedulerID && s.DeletedAt == nil {
			sched = &s
			break
		}
	}
	if sched == nil {
		return nil
	}

	nextRun := calculateNextRun(sched)

	// Update scheduler's next_run
	_ = h.Schedulers.WithLock(func(items []models.Scheduler) ([]models.Scheduler, error) {
		for i, s := range items {
			if s.ID == sched.ID {
				items[i].NextRun = nextRun
				items[i].UpdatedAt = time.Now().UTC()
				break
			}
		}
		return items, nil
	})

	// Get board's first swimlane (typically "To Do")
	boards, err := h.Boards.LoadAll()
	if err != nil {
		return nil
	}
	firstSwimlane := "To Do"
	for _, b := range boards {
		if b.ID == completed.BoardID && b.DeletedAt == nil && len(b.Swimlanes) > 0 {
			firstSwimlane = b.Swimlanes[0]
			break
		}
	}

	newTask := models.Task{
		ID:                uuid.New().String(),
		BoardID:           completed.BoardID,
		Swimlane:          firstSwimlane,
		TaskType:          completed.TaskType,
		Title:             completed.Title,
		Description:       completed.Description,
		AssigneeID:        completed.AssigneeID,
		EstimationMinutes: completed.EstimationMinutes,
		Cost:              completed.Cost,
		Priority:          completed.Priority,
		Reminders:         []models.Reminder{},
		SchedulerID:       completed.SchedulerID,
		DueDate:           nextRun,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	return &newTask
}
