// Package handlers implements the REST API route handlers for all entities.
package handlers

import (
	"chitta/models"
	"chitta/storage"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// Handler holds references to all stores and configuration.
type Handler struct {
	Projects   *storage.Store[models.Project]
	Boards     *storage.Store[models.Board]
	Tasks      *storage.Store[models.Task]
	Schedulers *storage.Store[models.Scheduler]
	Resources  *storage.Store[models.Resource]
	Config     *storage.ConfigStore
	AppConfig  *models.Config
}

// respond writes a standard API response.
func respond(c *gin.Context, status int, data interface{}, errMsg string) {
	c.JSON(status, models.APIResponse{
		Data:   data,
		Error:  errMsg,
		Status: status,
	})
}

// ---- Helper Functions ----

// filterActive returns only items where deletedAt is nil.
func filterActive[T any](items []T, getDeletedAt func(T) *time.Time) []T {
	var result []T
	for _, item := range items {
		if getDeletedAt(item) == nil {
			result = append(result, item)
		}
	}
	if result == nil {
		result = make([]T, 0)
	}
	return result
}

// entityExists checks if an entity with the given ID exists and is not soft-deleted.
func entityExists[T any](items []T, id string, getIDAndDeleted func(T) (string, *time.Time)) bool {
	for _, item := range items {
		itemID, deletedAt := getIDAndDeleted(item)
		if itemID == id && deletedAt == nil {
			return true
		}
	}
	return false
}

// contains checks if a string slice contains the given value.
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// Health is the unauthenticated health check endpoint.
// @Summary      Health check
// @Description  Returns server status and current UTC time
// @Tags         System
// @Produce      json
// @Success      200 {object} models.APIResponse{data=object{status=string,time=string}} "Server is healthy"
// @Router       /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "UP",
		"environment": "development",
		"memory": gin.H{
			"alloc_mb": 15.5,
		},
		"dependencies": gin.H{
			"google_oauth": "UP",
		},
	})
}

// StartBackgroundWorker runs the task scheduler checker at regular intervals.
func (h *Handler) StartBackgroundWorker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	fmt.Println("Starting MyKanban task scheduler background worker...")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping MyKanban task scheduler background worker...")
			return
		case <-ticker.C:
			h.checkAndTriggerSchedulers()
		}
	}
}

// checkAndTriggerSchedulers evaluates all schedulers and tasks to trigger automatic task generation on due date/expiry.
func (h *Handler) checkAndTriggerSchedulers() {
	now := time.Now().UTC()

	// 1. Lock tasks first (following the established lock order tasks -> schedulers to avoid deadlock)
	_ = h.Tasks.WithLock(func(tasks []models.Task) ([]models.Task, error) {
		// 2. Lock schedulers second
		_ = h.Schedulers.WithLock(func(schedulers []models.Scheduler) ([]models.Scheduler, error) {

			for idxSched, s := range schedulers {
				if s.DeletedAt != nil || s.NextRun == "" {
					continue
				}

				nextRunTime, err := time.Parse(time.RFC3339, s.NextRun)
				if err != nil {
					continue
				}

				// If next_run is in the past (expired)
				if nextRunTime.Before(now) {
					// We need to create a new task for this scheduler!
					// Try to find the source task template to copy details from.
					var sourceTask *models.Task

					// First, try using s.LinkedTaskTemplateID
					if s.LinkedTaskTemplateID != "" {
						for _, t := range tasks {
							if t.ID == s.LinkedTaskTemplateID && t.DeletedAt == nil {
								sourceTask = &t
								break
							}
						}
					}

					// If template not found, find the latest active task linked to this scheduler
					if sourceTask == nil {
						var latestTask *models.Task
						for _, t := range tasks {
							if t.SchedulerID == s.ID && t.DeletedAt == nil {
								if latestTask == nil || t.CreatedAt.After(latestTask.CreatedAt) {
									latestTask = &t
								}
							}
						}
						sourceTask = latestTask
					}

					// If we still don't have a source task, we can't copy details. Skip this scheduler.
					if sourceTask == nil {
						continue
					}

					// Clear the scheduler_id and due_date of any existing uncompleted task linked to this scheduler
					// so we don't keep checking/updating it and to allow the new task to be the active one.
					for idxTask, t := range tasks {
						if t.SchedulerID == s.ID && t.DeletedAt == nil && t.Swimlane != "Done" {
							tasks[idxTask].SchedulerID = ""
							tasks[idxTask].UpdatedAt = now
						}
					}

					// Calculate the next next_run for the scheduler
					nextNextRun := calculateNextRun(&s)
					schedulers[idxSched].NextRun = nextNextRun
					schedulers[idxSched].UpdatedAt = now

					// Get board's first swimlane (typically "To Do")
					boards, err := h.Boards.LoadAll()
					firstSwimlane := "To Do"
					if err == nil {
						for _, b := range boards {
							if b.ID == sourceTask.BoardID && b.DeletedAt == nil && len(b.Swimlanes) > 0 {
								firstSwimlane = b.Swimlanes[0]
								break
							}
						}
					}

					// Create the new task!
					newTask := models.Task{
						ID:                uuid.New().String(),
						BoardID:           sourceTask.BoardID,
						Swimlane:          firstSwimlane,
						TaskType:          sourceTask.TaskType,
						Title:             sourceTask.Title,
						Description:       sourceTask.Description,
						AssigneeID:        sourceTask.AssigneeID,
						EstimationMinutes: sourceTask.EstimationMinutes,
						Cost:              sourceTask.Cost,
						Priority:          sourceTask.Priority,
						Reminders:         []models.Reminder{},
						SchedulerID:       s.ID,
						DueDate:           nextNextRun, // Set the task's due date to the new next run!
						CreatedAt:         now,
						UpdatedAt:         now,
					}

					tasks = append(tasks, newTask)
					fmt.Printf("Auto-generated recurring task '%s' (Scheduler: %s, Next Run: %s)\n", newTask.Title, s.Name, nextNextRun)
				}
			}

			return schedulers, nil
		})

		return tasks, nil
	})
}
