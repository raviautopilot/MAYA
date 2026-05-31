package handlers

import (
	"chitta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// ---- Project Handlers ----

// CreateProject creates a new project.
// @Summary      Create project
// @Description  Create a new project (personal or professional)
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.Project true "Project data (name, description, type required)"
// @Success      201 {object} models.APIResponse{data=models.Project} "Created project"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/projects [post]
func (h *Handler) CreateProject(c *gin.Context) {
	var p models.Project
	if err := c.ShouldBindJSON(&p); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt

	if err := h.Projects.WithLock(func(items []models.Project) ([]models.Project, error) {
		return append(items, p), nil
	}); err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to save project: "+err.Error())
		return
	}
	respond(c, http.StatusCreated, p, "")
}

// ListProjects returns all active (non-deleted) projects.
// @Summary      List projects
// @Description  Get all active projects (excludes soft-deleted)
// @Tags         Projects
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse{data=[]models.Project} "List of projects"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/projects [get]
func (h *Handler) ListProjects(c *gin.Context) {
	items, err := h.Projects.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to load projects: "+err.Error())
		return
	}
	active := filterActive(items, func(p models.Project) *time.Time { return p.DeletedAt })
	respond(c, http.StatusOK, active, "")
}

// GetProject retrieves a single project by ID.
// @Summary      Get project
// @Description  Get a project by its UUID
// @Tags         Projects
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Success      200 {object} models.APIResponse{data=models.Project} "Project details"
// @Failure      404 {object} models.APIResponse "Project not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/projects/{id} [get]
func (h *Handler) GetProject(c *gin.Context) {
	id := c.Param("id")
	items, err := h.Projects.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to load projects: "+err.Error())
		return
	}
	for _, p := range items {
		if p.ID == id && p.DeletedAt == nil {
			respond(c, http.StatusOK, p, "")
			return
		}
	}
	respond(c, http.StatusNotFound, nil, "project not found")
}

// UpdateProject replaces a project's data.
// @Summary      Update project
// @Description  Full update of a project by UUID
// @Tags         Projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string        true "Project UUID"
// @Param        body body models.Project true "Updated project data"
// @Success      200 {object} models.APIResponse{data=models.Project} "Updated project"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      404 {object} models.APIResponse "Project not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/projects/{id} [put]
func (h *Handler) UpdateProject(c *gin.Context) {
	id := c.Param("id")
	var input models.Project
	if err := c.ShouldBindJSON(&input); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}

	var updated *models.Project
	err := h.Projects.WithLock(func(items []models.Project) ([]models.Project, error) {
		for i, p := range items {
			if p.ID == id && p.DeletedAt == nil {
				items[i].Name = input.Name
				items[i].Description = input.Description
				items[i].Type = input.Type
				items[i].StartDate = input.StartDate
				items[i].EndDate = input.EndDate
				items[i].EstimatedHours = input.EstimatedHours
				items[i].UpdatedAt = time.Now().UTC()
				updated = &items[i]
				return items, nil
			}
		}
		return items, fmt.Errorf("not found")
	})
	if err != nil {
		if err.Error() == "not found" {
			respond(c, http.StatusNotFound, nil, "project not found")
		} else {
			respond(c, http.StatusInternalServerError, nil, "failed to update project: "+err.Error())
		}
		return
	}
	respond(c, http.StatusOK, updated, "")
}

// DeleteProject soft-deletes a project.
// @Summary      Delete project
// @Description  Soft-delete a project by UUID (sets deleted_at timestamp)
// @Tags         Projects
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Project UUID"
// @Success      200 {object} models.APIResponse{data=object{message=string}} "Deletion confirmation"
// @Failure      404 {object} models.APIResponse "Project not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/projects/{id} [delete]
func (h *Handler) DeleteProject(c *gin.Context) {
	id := c.Param("id")

	// Verify project contains no associated active boards
	boards, err := h.Boards.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to check associated boards: "+err.Error())
		return
	}
	for _, b := range boards {
		if b.ProjectID == id && b.DeletedAt == nil {
			respond(c, http.StatusBadRequest, nil, "cannot delete project: associated boards exist")
			return
		}
	}

	now := time.Now().UTC()
	var deleted bool
	err = h.Projects.WithLock(func(items []models.Project) ([]models.Project, error) {
		for i, p := range items {
			if p.ID == id && p.DeletedAt == nil {
				items[i].DeletedAt = &now
				items[i].UpdatedAt = now
				deleted = true
				return items, nil
			}
		}
		return items, nil
	})
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to delete project: "+err.Error())
		return
	}
	if !deleted {
		respond(c, http.StatusNotFound, nil, "project not found")
		return
	}
	respond(c, http.StatusOK, gin.H{"message": "project deleted"}, "")
}
