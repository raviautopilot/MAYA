package handlers

import (
	"chitta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// ---- Board Handlers ----

// CreateBoard creates a new board within a project.
// @Summary      Create board
// @Description  Create a new kanban board linked to a project. Defaults swimlanes to [To Do, In Progress, Done] and task types to [Bug, Feature, Chore] if not provided.
// @Tags         Boards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.Board true "Board data (project_id, name required)"
// @Success      201 {object} models.APIResponse{data=models.Board} "Created board"
// @Failure      400 {object} models.APIResponse "Invalid request or invalid project_id"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/boards [post]
func (h *Handler) CreateBoard(c *gin.Context) {
	var b models.Board
	if err := c.ShouldBindJSON(&b); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}
	// Verify project exists
	projects, err := h.Projects.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to verify project: "+err.Error())
		return
	}
	if !entityExists(projects, b.ProjectID, func(p models.Project) (string, *time.Time) { return p.ID, p.DeletedAt }) {
		respond(c, http.StatusBadRequest, nil, "project_id does not reference a valid project")
		return
	}

	if len(b.Swimlanes) == 0 {
		b.Swimlanes = models.DefaultSwimlanes
	}
	if len(b.TaskTypes) == 0 {
		b.TaskTypes = models.DefaultTaskTypes
	}
	b.ID = uuid.New().String()
	b.IsActive = true
	b.CreatedAt = time.Now().UTC()
	b.UpdatedAt = b.CreatedAt

	if err := h.Boards.WithLock(func(items []models.Board) ([]models.Board, error) {
		return append(items, b), nil
	}); err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to save board: "+err.Error())
		return
	}
	respond(c, http.StatusCreated, b, "")
}

// ListBoards returns all active boards, optionally filtered by project_id.
// @Summary      List boards
// @Description  Get all active boards. Optionally filter by project_id query parameter.
// @Tags         Boards
// @Produce      json
// @Security     BearerAuth
// @Param        project_id query string false "Filter by project UUID"
// @Success      200 {object} models.APIResponse{data=[]models.Board} "List of boards"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/boards [get]
func (h *Handler) ListBoards(c *gin.Context) {
	items, err := h.Boards.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to load boards: "+err.Error())
		return
	}
	active := filterActive(items, func(b models.Board) *time.Time { return b.DeletedAt })

	// Optional filter by project_id
	if pid := c.Query("project_id"); pid != "" {
		var filtered []models.Board
		for _, b := range active {
			if b.ProjectID == pid {
				filtered = append(filtered, b)
			}
		}
		active = filtered
	}
	if active == nil {
		active = []models.Board{}
	}
	respond(c, http.StatusOK, active, "")
}

// GetBoard retrieves a single board by ID.
// @Summary      Get board
// @Description  Get a board by its UUID
// @Tags         Boards
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Board UUID"
// @Success      200 {object} models.APIResponse{data=models.Board} "Board details"
// @Failure      404 {object} models.APIResponse "Board not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/boards/{id} [get]
func (h *Handler) GetBoard(c *gin.Context) {
	id := c.Param("id")
	items, err := h.Boards.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to load boards: "+err.Error())
		return
	}
	for _, b := range items {
		if b.ID == id && b.DeletedAt == nil {
			respond(c, http.StatusOK, b, "")
			return
		}
	}
	respond(c, http.StatusNotFound, nil, "board not found")
}

// UpdateBoard replaces a board's data.
// @Summary      Update board
// @Description  Full update of a board by UUID
// @Tags         Boards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string       true "Board UUID"
// @Param        body body models.Board  true "Updated board data"
// @Success      200 {object} models.APIResponse{data=models.Board} "Updated board"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      404 {object} models.APIResponse "Board not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/boards/{id} [put]
func (h *Handler) UpdateBoard(c *gin.Context) {
	id := c.Param("id")
	var input models.Board
	if err := c.ShouldBindJSON(&input); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}

	var updated *models.Board
	err := h.Boards.WithLock(func(items []models.Board) ([]models.Board, error) {
		for i, b := range items {
			if b.ID == id && b.DeletedAt == nil {
				items[i].Name = input.Name
				items[i].ProjectID = input.ProjectID
				if len(input.Swimlanes) > 0 {
					items[i].Swimlanes = input.Swimlanes
				}
				if len(input.TaskTypes) > 0 {
					items[i].TaskTypes = input.TaskTypes
				}
				items[i].IsActive = input.IsActive
				items[i].UpdatedAt = time.Now().UTC()
				updated = &items[i]
				return items, nil
			}
		}
		return items, fmt.Errorf("not found")
	})
	if err != nil {
		if err.Error() == "not found" {
			respond(c, http.StatusNotFound, nil, "board not found")
		} else {
			respond(c, http.StatusInternalServerError, nil, "failed to update board: "+err.Error())
		}
		return
	}
	respond(c, http.StatusOK, updated, "")
}

// DeleteBoard soft-deletes a board.
// @Summary      Delete board
// @Description  Soft-delete a board by UUID
// @Tags         Boards
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Board UUID"
// @Success      200 {object} models.APIResponse{data=object{message=string}} "Deletion confirmation"
// @Failure      404 {object} models.APIResponse "Board not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/boards/{id} [delete]
func (h *Handler) DeleteBoard(c *gin.Context) {
	id := c.Param("id")

	// Verify board contains no associated active tasks
	tasks, err := h.Tasks.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to check associated tasks: "+err.Error())
		return
	}
	for _, t := range tasks {
		if t.BoardID == id && t.DeletedAt == nil {
			respond(c, http.StatusBadRequest, nil, "cannot delete board: associated tasks exist")
			return
		}
	}

	now := time.Now().UTC()
	var deleted bool
	err = h.Boards.WithLock(func(items []models.Board) ([]models.Board, error) {
		for i, b := range items {
			if b.ID == id && b.DeletedAt == nil {
				items[i].DeletedAt = &now
				items[i].UpdatedAt = now
				deleted = true
				return items, nil
			}
		}
		return items, nil
	})
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to delete board: "+err.Error())
		return
	}
	if !deleted {
		respond(c, http.StatusNotFound, nil, "board not found")
		return
	}
	respond(c, http.StatusOK, gin.H{"message": "board deleted"}, "")
}
