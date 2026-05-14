// Package handlers implements the REST API route handlers for all entities.
package handlers

import (
        "fmt"
        "net/http"
        "time"

        "github.com/gin-gonic/gin"
        "github.com/google/uuid"
        "github.com/robfig/cron/v3"

        "mykanban-backend/auth"
        "mykanban-backend/models"
        "mykanban-backend/storage"
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

// ---- Auth Handlers ----

// Login authenticates with email/password and returns a JWT.
// @Summary      User login
// @Description  Authenticate with email and password to receive a JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body models.LoginRequest true "Login credentials"
// @Success      200 {object} models.APIResponse{data=object{token=string}} "JWT token"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      401 {object} models.APIResponse "Invalid credentials"
// @Failure      500 {object} models.APIResponse "Token generation failure"
// @Router       /v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
        var req models.LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
                return
        }
        if req.Email != h.AppConfig.RootEmail {
                respond(c, http.StatusUnauthorized, nil, "invalid credentials")
                return
        }
        if err := auth.CheckPassword(h.AppConfig.RootPasswordHash, req.Password); err != nil {
                respond(c, http.StatusUnauthorized, nil, "invalid credentials")
                return
        }
        token, err := auth.GenerateJWT(req.Email, h.AppConfig.JWTSecret, h.AppConfig.JWTExpiryHours)
        if err != nil {
                respond(c, http.StatusInternalServerError, nil, "failed to generate token")
                return
        }
        respond(c, http.StatusOK, gin.H{"token": token}, "")
}

// ChangePassword updates the root password.
// @Summary      Change password
// @Description  Change the root user's password (requires current password)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body models.ChangePasswordRequest true "Password change request"
// @Success      200 {object} models.APIResponse{data=object{message=string}} "Password changed"
// @Failure      400 {object} models.APIResponse "Invalid request"
// @Failure      401 {object} models.APIResponse "Old password incorrect"
// @Failure      500 {object} models.APIResponse "Server error"
// @Router       /v1/auth/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
        var req models.ChangePasswordRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                respond(c, http.StatusBadRequest, nil, "invalid request: "+err.Error())
                return
        }
        if err := auth.CheckPassword(h.AppConfig.RootPasswordHash, req.OldPassword); err != nil {
                respond(c, http.StatusUnauthorized, nil, "old password is incorrect")
                return
        }
        newHash, err := auth.HashPassword(req.NewPassword)
        if err != nil {
                respond(c, http.StatusInternalServerError, nil, "failed to hash new password")
                return
        }
        h.AppConfig.RootPasswordHash = newHash
        if err := h.Config.Save(h.AppConfig); err != nil {
                respond(c, http.StatusInternalServerError, nil, "failed to save config")
                return
        }
        respond(c, http.StatusOK, gin.H{"message": "password changed successfully"}, "")
}

// GoogleLogin redirects to Google OAuth consent screen.
// @Summary      Google OAuth login
// @Description  Redirect to Google OAuth 2.0 consent screen for authentication
// @Tags         Auth
// @Produce      json
// @Success      307 "Redirect to Google"
// @Failure      503 {object} models.APIResponse "Google OAuth not configured"
// @Router       /v1/auth/google/login [get]
func (h *Handler) GoogleLogin(c *gin.Context) {
        if h.AppConfig.GoogleClientID == "" {
                respond(c, http.StatusServiceUnavailable, nil, "google oauth not configured")
                return
        }
        cfg := auth.GoogleOAuthConfig(h.AppConfig.GoogleClientID, h.AppConfig.GoogleClientSecret, h.AppConfig.GoogleRedirectURL)
        url := cfg.AuthCodeURL("state-token")
        c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the OAuth callback from Google.
// @Summary      Google OAuth callback
// @Description  Handle the callback from Google OAuth and return a JWT token
// @Tags         Auth
// @Produce      json
// @Param        code query string true "Authorization code from Google"
// @Success      200 {object} models.APIResponse{data=object{token=string,email=string,name=string}} "JWT token and user info"
// @Failure      400 {object} models.APIResponse "Missing code parameter"
// @Failure      500 {object} models.APIResponse "Google auth or token generation failure"
// @Router       /v1/auth/google/callback [get]
func (h *Handler) GoogleCallback(c *gin.Context) {
        code := c.Query("code")
        if code == "" {
                respond(c, http.StatusBadRequest, nil, "missing code parameter")
                return
        }
        cfg := auth.GoogleOAuthConfig(h.AppConfig.GoogleClientID, h.AppConfig.GoogleClientSecret, h.AppConfig.GoogleRedirectURL)
        userInfo, err := auth.FetchGoogleUser(c.Request.Context(), cfg, code)
        if err != nil {
                respond(c, http.StatusInternalServerError, nil, "google auth failed: "+err.Error())
                return
        }
        token, err := auth.GenerateJWT(userInfo.Email, h.AppConfig.JWTSecret, h.AppConfig.JWTExpiryHours)
        if err != nil {
                respond(c, http.StatusInternalServerError, nil, "failed to generate token")
                return
        }
        respond(c, http.StatusOK, gin.H{"token": token, "email": userInfo.Email, "name": userInfo.Name}, "")
}

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
        now := time.Now().UTC()
        var deleted bool
        err := h.Projects.WithLock(func(items []models.Project) ([]models.Project, error) {
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
        now := time.Now().UTC()
        var deleted bool
        err := h.Boards.WithLock(func(items []models.Board) ([]models.Board, error) {
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
                CreatedAt:         time.Now().UTC(),
                UpdatedAt:         time.Now().UTC(),
        }
        return &newTask
}

// calculateNextRun determines the next occurrence based on the scheduler type/cron.
func calculateNextRun(s *models.Scheduler) string {
        now := time.Now().UTC()
        switch s.Type {
        case "daily":
                return now.AddDate(0, 0, 1).Format(time.RFC3339)
        case "weekly":
                return now.AddDate(0, 0, 7).Format(time.RFC3339)
        case "monthly":
                return now.AddDate(0, 1, 0).Format(time.RFC3339)
        case "yearly":
                return now.AddDate(1, 0, 0).Format(time.RFC3339)
        case "cron":
                if s.CronExpression != "" {
                        parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
                        schedule, err := parser.Parse(s.CronExpression)
                        if err == nil {
                                return schedule.Next(now).Format(time.RFC3339)
                        }
                }
                return now.AddDate(0, 0, 1).Format(time.RFC3339)
        default:
                return now.AddDate(0, 0, 1).Format(time.RFC3339)
        }
}

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
        respond(c, http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)}, "")
}
