package handlers

import (
	"chitta/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"net/http"
	"time"
)

// ---- Scheduler Handlers ----

// CreateScheduler creates a new recurring schedule.
// @Summary      Create scheduler
// @Description  Create a new scheduler (daily, weekly, monthly, yearly, or cron). Automatically calculates next_run if not provided.
// @Tags         Schedulers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.Scheduler true "Scheduler data (name, type required; cron_expression required when type=cron)"
// @Success      201 {object} models.APIResponse{data=models.Scheduler} "Created scheduler"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/schedulers [post]
func (h *Handler) CreateScheduler(c *gin.Context) {
	var s models.Scheduler
	if err := c.ShouldBindJSON(&s); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}
	s.ID = uuid.New().String()
	s.CreatedAt = time.Now().UTC()
	s.UpdatedAt = s.CreatedAt
	if s.NextRun == "" {
		s.NextRun = calculateNextRun(&s)
	}

	if err := h.Schedulers.WithLock(func(items []models.Scheduler) ([]models.Scheduler, error) {
		return append(items, s), nil
	}); err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to save scheduler: "+err.Error())
		return
	}
	respond(c, http.StatusCreated, s, "")
}

// ListSchedulers returns all active schedulers.
// @Summary      List schedulers
// @Description  Get all active schedulers (excludes soft-deleted)
// @Tags         Schedulers
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse{data=[]models.Scheduler} "List of schedulers"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/schedulers [get]
func (h *Handler) ListSchedulers(c *gin.Context) {
	items, err := h.Schedulers.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to load schedulers: "+err.Error())
		return
	}
	active := filterActive(items, func(s models.Scheduler) *time.Time { return s.DeletedAt })
	respond(c, http.StatusOK, active, "")
}

// GetScheduler retrieves a single scheduler by ID.
// @Summary      Get scheduler
// @Description  Get a scheduler by its UUID
// @Tags         Schedulers
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Scheduler UUID"
// @Success      200 {object} models.APIResponse{data=models.Scheduler} "Scheduler details"
// @Failure      404 {object} models.APIResponse "Scheduler not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/schedulers/{id} [get]
func (h *Handler) GetScheduler(c *gin.Context) {
	id := c.Param("id")
	items, err := h.Schedulers.LoadAll()
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to load schedulers: "+err.Error())
		return
	}
	for _, s := range items {
		if s.ID == id && s.DeletedAt == nil {
			respond(c, http.StatusOK, s, "")
			return
		}
	}
	respond(c, http.StatusNotFound, nil, "scheduler not found")
}

// UpdateScheduler replaces a scheduler's data.
// @Summary      Update scheduler
// @Description  Full update of a scheduler by UUID. Recalculates next_run.
// @Tags         Schedulers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path string           true "Scheduler UUID"
// @Param        body body models.Scheduler  true "Updated scheduler data"
// @Success      200 {object} models.APIResponse{data=models.Scheduler} "Updated scheduler"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      404 {object} models.APIResponse "Scheduler not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/schedulers/{id} [put]
func (h *Handler) UpdateScheduler(c *gin.Context) {
	id := c.Param("id")
	var input models.Scheduler
	if err := c.ShouldBindJSON(&input); err != nil {
		respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
		return
	}

	var updated *models.Scheduler
	err := h.Schedulers.WithLock(func(items []models.Scheduler) ([]models.Scheduler, error) {
		for i, s := range items {
			if s.ID == id && s.DeletedAt == nil {
				items[i].Name = input.Name
				items[i].CronExpression = input.CronExpression
				items[i].Type = input.Type
				items[i].LinkedTaskTemplateID = input.LinkedTaskTemplateID
				items[i].NextRun = calculateNextRun(&input)
				items[i].UpdatedAt = time.Now().UTC()
				updated = &items[i]
				return items, nil
			}
		}
		return items, fmt.Errorf("not found")
	})
	if err != nil {
		if err.Error() == "not found" {
			respond(c, http.StatusNotFound, nil, "scheduler not found")
		} else {
			respond(c, http.StatusInternalServerError, nil, "failed to update scheduler: "+err.Error())
		}
		return
	}
	respond(c, http.StatusOK, updated, "")
}

// DeleteScheduler soft-deletes a scheduler.
// @Summary      Delete scheduler
// @Description  Soft-delete a scheduler by UUID
// @Tags         Schedulers
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Scheduler UUID"
// @Success      200 {object} models.APIResponse{data=object{message=string}} "Deletion confirmation"
// @Failure      404 {object} models.APIResponse "Scheduler not found"
// @Failure      500 {object} models.APIResponse "Storage error"
// @Router       /v1/schedulers/{id} [delete]
func (h *Handler) DeleteScheduler(c *gin.Context) {
	id := c.Param("id")
	now := time.Now().UTC()
	var deleted bool
	err := h.Schedulers.WithLock(func(items []models.Scheduler) ([]models.Scheduler, error) {
		for i, s := range items {
			if s.ID == id && s.DeletedAt == nil {
				items[i].DeletedAt = &now
				items[i].UpdatedAt = now
				deleted = true
				return items, nil
			}
		}
		return items, nil
	})
	if err != nil {
		respond(c, http.StatusInternalServerError, nil, "failed to delete scheduler: "+err.Error())
		return
	}
	if !deleted {
		respond(c, http.StatusNotFound, nil, "scheduler not found")
		return
	}
	respond(c, http.StatusOK, gin.H{"message": "scheduler deleted"}, "")
}

// calculateNextRun determines the next occurrence based on the scheduler type/cron.
func calculateNextRun(s *models.Scheduler) string {
	now := time.Now().UTC()
	var baseTime time.Time
	if s.NextRun != "" {
		if t, err := time.Parse(time.RFC3339, s.NextRun); err == nil {
			baseTime = t
		}
	}

	if baseTime.IsZero() {
		baseTime = now
	}

	switch s.Type {
	case "daily":
		next := baseTime.AddDate(0, 0, 1)
		for next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}
		return next.Format(time.RFC3339)
	case "weekly":
		next := baseTime.AddDate(0, 0, 7)
		for next.Before(now) {
			next = next.AddDate(0, 0, 7)
		}
		return next.Format(time.RFC3339)
	case "monthly":
		next := baseTime.AddDate(0, 1, 0)
		for next.Before(now) {
			next = next.AddDate(0, 1, 0)
		}
		return next.Format(time.RFC3339)
	case "yearly":
		next := baseTime.AddDate(1, 0, 0)
		for next.Before(now) {
			next = next.AddDate(1, 0, 0)
		}
		return next.Format(time.RFC3339)
	case "cron":
		if s.CronExpression != "" {
			parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			schedule, err := parser.Parse(s.CronExpression)
			if err == nil {
				next := schedule.Next(baseTime)
				for next.Before(now) {
					next = schedule.Next(next)
				}
				return next.Format(time.RFC3339)
			}
		}
		next := baseTime.AddDate(0, 0, 1)
		for next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}
		return next.Format(time.RFC3339)
	default:
		next := baseTime.AddDate(0, 0, 1)
		for next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}
		return next.Format(time.RFC3339)
	}
}
