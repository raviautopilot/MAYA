// Package models defines all entity data structures for the MyKanban application.
package models

import "time"

// APIResponse is the standard JSON envelope for all API responses.
type APIResponse struct {
	Data   interface{} `json:"data"`
	Error  string      `json:"error"`
	Status int         `json:"status"`
}

// Project represents a top-level project grouping boards.
type Project struct {
	ID          string     `json:"id"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	Type        string     `json:"type" binding:"required,oneof=personal professional"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// Board represents a kanban board within a project.
type Board struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"project_id" binding:"required"`
	Name      string     `json:"name" binding:"required"`
	Swimlanes []string   `json:"swimlanes"`
	TaskTypes []string   `json:"task_types"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// DefaultSwimlanes are the default swimlane columns for a board.
var DefaultSwimlanes = []string{"To Do", "In Progress", "Done"}

// DefaultTaskTypes are the default task type categories for a board.
var DefaultTaskTypes = []string{"Bug", "Feature", "Chore"}

// Reminder represents a timed reminder attached to a task.
type Reminder struct {
	Time string `json:"time" binding:"required"`
	Note string `json:"note"`
}

// Task represents a work item on a board.
type Task struct {
	ID                string     `json:"id"`
	BoardID           string     `json:"board_id" binding:"required"`
	Swimlane          string     `json:"swimlane" binding:"required"`
	TaskType          string     `json:"task_type" binding:"required"`
	Title             string     `json:"title" binding:"required"`
	Description       string     `json:"description"`
	AssigneeID        string     `json:"assignee_id,omitempty"`
	EstimationMinutes int        `json:"estimation_minutes"`
	ActualTimeMinutes int        `json:"actual_time_minutes"`
	Cost              float64    `json:"cost"`
	Priority          string     `json:"priority" binding:"required,oneof=Low Medium High Critical"`
	Reminders         []Reminder `json:"reminders"`
	SchedulerID       string     `json:"scheduler_id,omitempty"`
	DueDate           string     `json:"due_date,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}

// MaxReminders is the maximum number of reminders allowed per task.
const MaxReminders = 5

// Scheduler represents a recurring schedule definition.
type Scheduler struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name" binding:"required"`
	CronExpression       string     `json:"cron_expression"`
	Type                 string     `json:"type" binding:"required,oneof=cron yearly monthly weekly daily"`
	NextRun              string     `json:"next_run"`
	LinkedTaskTemplateID string     `json:"linked_task_template_id,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`
}

// Resource represents a person or service that can be linked to entities.
type Resource struct {
	ID          string     `json:"id"`
	Name        string     `json:"name" binding:"required"`
	Type        string     `json:"type" binding:"required,oneof=Global Project Task"`
	LinkedItems []string   `json:"linked_items"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// Config holds the application configuration loaded from config.json.
type Config struct {
	Port               int    `json:"port"`
	RootEmail          string `json:"root_email"`
	RootPasswordHash   string `json:"root_password_hash"`
	JWTSecret          string `json:"jwt_secret"`
	JWTExpiryHours     int    `json:"jwt_expiry_hours"`
	GoogleClientID     string `json:"google_client_id"`
	GoogleClientSecret string `json:"google_client_secret"`
	GoogleRedirectURL  string `json:"google_redirect_url"`
	StorageDir         string `json:"storage_dir"`
	LogFile            string `json:"log_file"`
	AllowedOrigins     string `json:"allowed_origins"`
}

// LoginRequest is the payload for email/password login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest is the payload for changing the root password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
