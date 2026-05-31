package handlers

import (
	"chitta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// ---- Resource Handlers ----

// CreateResource creates a new resource.
// @Summary      Create resource
// @Description  Create a new resource (Global, Project, or Task scoped)
// @Tags         Resources
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.Resource true "Resource data (name, type required)"
// @Success      201 {object} models.APIResponse{data=models.Resource} "Created resource"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/resources [post]
func (h *Handler) CreateResource(c *gin.Context) {
	var r models.Resource
	if err := c.ShouldBindJSON(&r); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}
	r.ID = uuid.New().String()
	r.CreatedAt = time.Now().UTC()
	r.UpdatedAt = r.CreatedAt
	if r.LinkedItems == nil {
		r.LinkedItems = []string{}
	}

	if err := h.Resources.WithLock(func(items []models.Resource) ([]models.Resource, error) {
		return append(items, r), nil
	}); err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to save resource: "+err.Error())
		return
	}
	respond(c, http.StatusCreated, r, "")
}

// ListResources returns all active resources.
// @Summary      List resources
// @Description  Get all active resources (excludes soft-deleted)
// @Tags         Resources
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse{data=[]models.Resource} "List of resources"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/resources [get]
func (h *Handler) ListResources(c *gin.Context) {
	items, err := h.Resources.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to load resources: "+err.Error())
		return
	}
	active := filterActive(items, func(r models.Resource) *time.Time { return r.DeletedAt })
	respond(c, http.StatusOK, active, "")
}

// GetResource retrieves a single resource by ID.
// @Summary      Get resource
// @Description  Get a resource by its UUID
// @Tags         Resources
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Resource UUID"
// @Success      200 {object} models.APIResponse{data=models.Resource} "Resource details"
// @Failure      404 {object} models.APIResponse "Resource not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/resources/{id} [get]
func (h *Handler) GetResource(c *gin.Context) {
	id := c.Param("id")
	items, err := h.Resources.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to load resources: "+err.Error())
		return
	}
	for _, r := range items {
		if r.ID == id && r.DeletedAt == nil {
			respond(c, http.StatusOK, r, "")
			return
		}
	}
	respond(c, http.StatusNotFound, nil, "resource not found")
}

// UpdateResource replaces a resource's data.
// @Summary      Update resource
// @Description  Full update of a resource by UUID
// @Tags         Resources
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string          true "Resource UUID"
// @Param        body body models.Resource  true "Updated resource data"
// @Success      200 {object} models.APIResponse{data=models.Resource} "Updated resource"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      404 {object} models.APIResponse "Resource not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/resources/{id} [put]
func (h *Handler) UpdateResource(c *gin.Context) {
	id := c.Param("id")
	var input models.Resource
	if err := c.ShouldBindJSON(&input); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}

	var updated *models.Resource
	err := h.Resources.WithLock(func(items []models.Resource) ([]models.Resource, error) {
		for i, r := range items {
			if r.ID == id && r.DeletedAt == nil {
				items[i].Name = input.Name
				items[i].Type = input.Type
				if input.LinkedItems != nil {
					items[i].LinkedItems = input.LinkedItems
				}
				items[i].ResourceType = input.ResourceType
				items[i].ResourceRole = input.ResourceRole
				items[i].HourlyRate = input.HourlyRate
				items[i].DailyRate = input.DailyRate
				items[i].UpdatedAt = time.Now().UTC()
				updated = &items[i]
				return items, nil
			}
		}
		return items, fmt.Errorf("not found")
	})
	if err != nil {
		if err.Error() == "not found" {
			respond(c, http.StatusNotFound, nil, "resource not found")
		} else {
			respond(c, http.StatusInternalServerError, nil, "failed to update resource: "+err.Error())
		}
		return
	}
	respond(c, http.StatusOK, updated, "")
}

// DeleteResource soft-deletes a resource.
// @Summary      Delete resource
// @Description  Soft-delete a resource by UUID
// @Tags         Resources
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Resource UUID"
// @Success      200 {object} models.APIResponse{data=object{message=string}} "Deletion confirmation"
// @Failure      404 {object} models.APIResponse "Resource not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/resources/{id} [delete]
func (h *Handler) DeleteResource(c *gin.Context) {
	id := c.Param("id")
	now := time.Now().UTC()
	var deleted bool
	err := h.Resources.WithLock(func(items []models.Resource) ([]models.Resource, error) {
		for i, r := range items {
			if r.ID == id && r.DeletedAt == nil {
				items[i].DeletedAt = &now
				items[i].UpdatedAt = now
				deleted = true
				return items, nil
			}
		}
		return items, nil
	})
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to delete resource: "+err.Error())
		return
	}
	if !deleted {
		respond(c, http.StatusNotFound, nil, "resource not found")
		return
	}
	respond(c, http.StatusOK, gin.H{"message": "resource deleted"}, "")
}
