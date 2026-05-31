package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"chitta/auth"
	"chitta/models"
	"chitta/storage"
)

// --- Mock FileSystem ---

type MockFS struct {
	mu    sync.RWMutex
	Files map[string][]byte
	Dirs  map[string]bool
}

func NewMockFS() *MockFS {
	return &MockFS{Files: make(map[string][]byte), Dirs: make(map[string]bool)}
}

func (m *MockFS) ReadFile(path string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.Files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return data, nil
}

func (m *MockFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Files[path] = data
	return nil
}

func (m *MockFS) MkdirAll(path string, perm os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Dirs[path] = true
	return nil
}

func (m *MockFS) Stat(path string) (os.FileInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.Files[path]; ok {
		return nil, nil
	}
	return nil, os.ErrNotExist
}

// --- Test Setup ---

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestHandler(t *testing.T) (*Handler, *gin.Engine, string) {
	t.Helper()
	fs := NewMockFS()

	projectStore, _ := storage.NewStore[models.Project]("/tmp/test", "projects.json", fs)
	boardStore, _ := storage.NewStore[models.Board]("/tmp/test", "boards.json", fs)
	taskStore, _ := storage.NewStore[models.Task]("/tmp/test", "tasks.json", fs)
	schedulerStore, _ := storage.NewStore[models.Scheduler]("/tmp/test", "schedulers.json", fs)
	resourceStore, _ := storage.NewStore[models.Resource]("/tmp/test", "resources.json", fs)

	pwHash, _ := auth.HashPassword("admin123")
	cfg := &models.Config{
		Port:             8080,
		RootEmail:        "admin@test.com",
		RootPasswordHash: pwHash,
		JWTSecret:        "test-jwt-secret",
		JWTExpiryHours:   24,
	}

	cfgData, _ := json.Marshal(cfg)
	fs.Files["/tmp/config.json"] = cfgData
	cfgStore := storage.NewConfigStore("/tmp/config.json", fs)

	h := &Handler{
		Projects:   projectStore,
		Boards:     boardStore,
		Tasks:      taskStore,
		Schedulers: schedulerStore,
		Resources:  resourceStore,
		Config:     cfgStore,
		AppConfig:  cfg,
	}

	router := gin.New()

	// Auth routes (public)
	router.POST("/api/v1/auth/login", h.Login)
	router.GET("/api/v1/auth/google/login", h.GoogleLogin)
	router.GET("/api/v1/auth/google/callback", h.GoogleCallback)

	// Protected routes
	api := router.Group("/api/v1")
	{
		api.POST("/auth/change-password", h.ChangePassword)
		api.POST("/projects", h.CreateProject)
		api.GET("/projects", h.ListProjects)
		api.GET("/projects/:id", h.GetProject)
		api.PUT("/projects/:id", h.UpdateProject)
		api.DELETE("/projects/:id", h.DeleteProject)
		api.POST("/boards", h.CreateBoard)
		api.GET("/boards", h.ListBoards)
		api.GET("/boards/:id", h.GetBoard)
		api.PUT("/boards/:id", h.UpdateBoard)
		api.DELETE("/boards/:id", h.DeleteBoard)
		api.POST("/tasks", h.CreateTask)
		api.GET("/tasks", h.ListTasks)
		api.GET("/tasks/:id", h.GetTask)
		api.PUT("/tasks/:id", h.UpdateTask)
		api.PATCH("/tasks/:id", h.PatchTask)
		api.DELETE("/tasks/:id", h.DeleteTask)
		api.POST("/schedulers", h.CreateScheduler)
		api.GET("/schedulers", h.ListSchedulers)
		api.GET("/schedulers/:id", h.GetScheduler)
		api.PUT("/schedulers/:id", h.UpdateScheduler)
		api.DELETE("/schedulers/:id", h.DeleteScheduler)
		api.POST("/resources", h.CreateResource)
		api.GET("/resources", h.ListResources)
		api.GET("/resources/:id", h.GetResource)
		api.PUT("/resources/:id", h.UpdateResource)
		api.DELETE("/resources/:id", h.DeleteResource)
	}

	router.GET("/api/health", Health)

	token, _ := auth.GenerateJWT("admin@test.com", "test-jwt-secret", 24)
	return h, router, token
}

func doRequest(router *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func parseResponse(w *httptest.ResponseRecorder) models.APIResponse {
	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

// --- Auth Tests ---

func TestLogin_Success(t *testing.T) {
	_, router, _ := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/auth/login", models.LoginRequest{
		Email: "admin@test.com", Password: "admin123",
	}, "")

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	resp := parseResponse(w)
	data := resp.Data.(map[string]interface{})
	if _, ok := data["token"]; !ok {
		t.Error("expected token in response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	_, router, _ := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/auth/login", models.LoginRequest{
		Email: "admin@test.com", Password: "wrong",
	}, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_WrongEmail(t *testing.T) {
	_, router, _ := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/auth/login", models.LoginRequest{
		Email: "wrong@test.com", Password: "admin123",
	}, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_InvalidBody(t *testing.T) {
	_, router, _ := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/auth/login", map[string]string{"bad": "body"}, "")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestChangePassword_Success(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/auth/change-password", models.ChangePasswordRequest{
		OldPassword: "admin123", NewPassword: "newpassword123",
	}, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestChangePassword_WrongOld(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/auth/change-password", models.ChangePasswordRequest{
		OldPassword: "wrongold", NewPassword: "newpassword123",
	}, token)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// --- Health Check ---

func TestHealth(t *testing.T) {
	_, router, _ := setupTestHandler(t)
	w := doRequest(router, "GET", "/api/health", nil, "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// --- Project CRUD Tests ---

func createProject(t *testing.T, router *gin.Engine, token string) map[string]interface{} {
	t.Helper()
	w := doRequest(router, "POST", "/api/v1/projects", map[string]interface{}{
		"name": "Test Project", "type": "personal", "description": "desc",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create project failed: %d %s", w.Code, w.Body.String())
	}
	resp := parseResponse(w)
	return resp.Data.(map[string]interface{})
}

func TestProject_Create(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	if p["name"] != "Test Project" {
		t.Errorf("expected 'Test Project', got '%v'", p["name"])
	}
	if p["id"] == "" {
		t.Error("expected non-empty ID")
	}
}

func TestProject_Create_InvalidType(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/projects", map[string]interface{}{
		"name": "Bad", "type": "invalid",
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestProject_List(t *testing.T) {
	_, router, token := setupTestHandler(t)
	createProject(t, router, token)
	createProject(t, router, token)

	w := doRequest(router, "GET", "/api/v1/projects", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	resp := parseResponse(w)
	items := resp.Data.([]interface{})
	if len(items) != 2 {
		t.Errorf("expected 2 projects, got %d", len(items))
	}
}

func TestProject_Get(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	id := p["id"].(string)

	w := doRequest(router, "GET", "/api/v1/projects/"+id, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestProject_Get_NotFound(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "GET", "/api/v1/projects/nonexistent", nil, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestProject_Update(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	id := p["id"].(string)

	w := doRequest(router, "PUT", "/api/v1/projects/"+id, map[string]interface{}{
		"name": "Updated", "type": "professional", "description": "new desc",
	}, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	resp := parseResponse(w)
	data := resp.Data.(map[string]interface{})
	if data["name"] != "Updated" {
		t.Errorf("expected 'Updated', got '%v'", data["name"])
	}
}

func TestProject_Delete(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	id := p["id"].(string)

	w := doRequest(router, "DELETE", "/api/v1/projects/"+id, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Verify soft-deleted (not in list)
	w2 := doRequest(router, "GET", "/api/v1/projects", nil, token)
	resp := parseResponse(w2)
	items := resp.Data.([]interface{})
	if len(items) != 0 {
		t.Errorf("expected 0 projects after delete, got %d", len(items))
	}
}

func TestProject_Delete_BlockedByBoards(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	pid := p["id"].(string)

	// Create board associated with project
	createBoard(t, router, token, pid)

	// Try deleting project (should fail with 400 Bad Request)
	w := doRequest(router, "DELETE", "/api/v1/projects/"+pid, nil, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
	resp := parseResponse(w)
	if resp.Error == "" {
		t.Error("expected error message in response, got empty")
	}
}

func TestProject_Delete_NotFound(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "DELETE", "/api/v1/projects/nonexistent", nil, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// --- Board CRUD Tests ---

func createBoard(t *testing.T, router *gin.Engine, token, projectID string) map[string]interface{} {
	t.Helper()
	w := doRequest(router, "POST", "/api/v1/boards", map[string]interface{}{
		"project_id": projectID, "name": "Test Board",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create board failed: %d %s", w.Code, w.Body.String())
	}
	resp := parseResponse(w)
	return resp.Data.(map[string]interface{})
}

func TestBoard_Create(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	if b["name"] != "Test Board" {
		t.Errorf("expected 'Test Board', got '%v'", b["name"])
	}
	// Check defaults
	swimlanes := b["swimlanes"].([]interface{})
	if len(swimlanes) != 3 {
		t.Errorf("expected 3 default swimlanes, got %d", len(swimlanes))
	}
}

func TestBoard_Create_CustomSwimlanes(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	w := doRequest(router, "POST", "/api/v1/boards", map[string]interface{}{
		"project_id": p["id"].(string), "name": "Custom",
		"swimlanes": []string{"A", "B"}, "task_types": []string{"X"},
	}, token)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	resp := parseResponse(w)
	data := resp.Data.(map[string]interface{})
	swimlanes := data["swimlanes"].([]interface{})
	if len(swimlanes) != 2 {
		t.Errorf("expected 2 swimlanes, got %d", len(swimlanes))
	}
}

func TestBoard_Create_InvalidProject(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/boards", map[string]interface{}{
		"project_id": "nonexistent", "name": "Bad Board",
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestBoard_List(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	pid := p["id"].(string)
	createBoard(t, router, token, pid)

	w := doRequest(router, "GET", "/api/v1/boards?project_id="+pid, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	resp := parseResponse(w)
	items := resp.Data.([]interface{})
	if len(items) != 1 {
		t.Errorf("expected 1 board, got %d", len(items))
	}
}

func TestBoard_GetUpdateDelete(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	bid := b["id"].(string)

	// Get
	w := doRequest(router, "GET", "/api/v1/boards/"+bid, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("GET expected 200, got %d", w.Code)
	}

	// Update
	w = doRequest(router, "PUT", "/api/v1/boards/"+bid, map[string]interface{}{
		"project_id": p["id"].(string), "name": "Updated Board",
	}, token)
	if w.Code != http.StatusOK {
		t.Errorf("PUT expected 200, got %d", w.Code)
	}

	// Delete
	w = doRequest(router, "DELETE", "/api/v1/boards/"+bid, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("DELETE expected 200, got %d", w.Code)
	}
}

func TestBoard_Delete_BlockedByTasks(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	bid := b["id"].(string)

	// Create task associated with board
	createTask(t, router, token, bid)

	// Try deleting board (should fail with 400 Bad Request)
	w := doRequest(router, "DELETE", "/api/v1/boards/"+bid, nil, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
	resp := parseResponse(w)
	if resp.Error == "" {
		t.Error("expected error message in response, got empty")
	}
}

// --- Task CRUD Tests ---

func createTask(t *testing.T, router *gin.Engine, token, boardID string) map[string]interface{} {
	t.Helper()
	w := doRequest(router, "POST", "/api/v1/tasks", map[string]interface{}{
		"board_id": boardID, "swimlane": "To Do", "task_type": "Feature",
		"title": "Test Task", "priority": "Medium",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create task failed: %d %s", w.Code, w.Body.String())
	}
	resp := parseResponse(w)
	return resp.Data.(map[string]interface{})
}

func TestTask_Create(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	task := createTask(t, router, token, b["id"].(string))
	if task["title"] != "Test Task" {
		t.Errorf("expected 'Test Task', got '%v'", task["title"])
	}
}

func TestTask_Create_InvalidSwimlane(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	w := doRequest(router, "POST", "/api/v1/tasks", map[string]interface{}{
		"board_id": b["id"].(string), "swimlane": "Invalid", "task_type": "Feature",
		"title": "Bad", "priority": "Medium",
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTask_Create_TooManyReminders(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	reminders := make([]map[string]string, 6)
	for i := range reminders {
		reminders[i] = map[string]string{"time": "2026-01-01T00:00:00Z", "note": "r"}
	}
	w := doRequest(router, "POST", "/api/v1/tasks", map[string]interface{}{
		"board_id": b["id"].(string), "swimlane": "To Do", "task_type": "Feature",
		"title": "Many Reminders", "priority": "Medium", "reminders": reminders,
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for >5 reminders, got %d", w.Code)
	}
}

func TestTask_List_Filters(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	bid := b["id"].(string)
	createTask(t, router, token, bid)

	// Filter by board_id
	w := doRequest(router, "GET", "/api/v1/tasks?board_id="+bid, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	resp := parseResponse(w)
	items := resp.Data.([]interface{})
	if len(items) != 1 {
		t.Errorf("expected 1 task, got %d", len(items))
	}

	// Filter by priority
	w = doRequest(router, "GET", "/api/v1/tasks?priority=High", nil, token)
	resp = parseResponse(w)
	items = resp.Data.([]interface{})
	if len(items) != 0 {
		t.Errorf("expected 0 tasks with High priority, got %d", len(items))
	}
}

func TestTask_PatchMove(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	task := createTask(t, router, token, b["id"].(string))
	tid := task["id"].(string)

	w := doRequest(router, "PATCH", "/api/v1/tasks/"+tid, map[string]interface{}{
		"swimlane": "In Progress",
	}, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	resp := parseResponse(w)
	data := resp.Data.(map[string]interface{})
	if data["swimlane"] != "In Progress" {
		t.Errorf("expected 'In Progress', got '%v'", data["swimlane"])
	}
}

func TestTask_UpdateAndDelete(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	task := createTask(t, router, token, b["id"].(string))
	tid := task["id"].(string)
	bid := b["id"].(string)

	// Full update
	w := doRequest(router, "PUT", "/api/v1/tasks/"+tid, map[string]interface{}{
		"board_id": bid, "swimlane": "In Progress", "task_type": "Bug",
		"title": "Updated Task", "priority": "High",
	}, token)
	if w.Code != http.StatusOK {
		t.Errorf("PUT expected 200, got %d", w.Code)
	}

	// Delete
	w = doRequest(router, "DELETE", "/api/v1/tasks/"+tid, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("DELETE expected 200, got %d", w.Code)
	}

	// Get after delete
	w = doRequest(router, "GET", "/api/v1/tasks/"+tid, nil, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
}

// --- Recurring Task Generation Test ---

func TestTask_RecurringGeneration(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	bid := b["id"].(string)

	// Create a scheduler
	w := doRequest(router, "POST", "/api/v1/schedulers", map[string]interface{}{
		"name": "Weekly", "type": "weekly",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create scheduler failed: %d", w.Code)
	}
	schedResp := parseResponse(w)
	schedData := schedResp.Data.(map[string]interface{})
	schedID := schedData["id"].(string)

	// Create a task linked to the scheduler
	w = doRequest(router, "POST", "/api/v1/tasks", map[string]interface{}{
		"board_id": bid, "swimlane": "To Do", "task_type": "Chore",
		"title": "Recurring Task", "priority": "Medium", "scheduler_id": schedID,
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create task failed: %d", w.Code)
	}
	taskResp := parseResponse(w)
	taskData := taskResp.Data.(map[string]interface{})
	taskID := taskData["id"].(string)

	// Move to Done -> should trigger recurring task creation
	w = doRequest(router, "PATCH", "/api/v1/tasks/"+taskID, map[string]interface{}{
		"swimlane": "Done",
	}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("patch to Done failed: %d", w.Code)
	}

	// List tasks - should have 2 (original Done + new To Do)
	w = doRequest(router, "GET", "/api/v1/tasks?board_id="+bid, nil, token)
	resp := parseResponse(w)
	items := resp.Data.([]interface{})
	if len(items) != 2 {
		t.Errorf("expected 2 tasks after recurring generation, got %d", len(items))
	}

	// Verify new task is in "To Do"
	var foundNewTodo bool
	for _, item := range items {
		m := item.(map[string]interface{})
		if m["id"] != taskID && m["swimlane"] == "To Do" && m["title"] == "Recurring Task" {
			foundNewTodo = true
		}
	}
	if !foundNewTodo {
		t.Error("expected a new recurring task in 'To Do' swimlane")
	}
}

// --- Scheduler CRUD Tests ---

func TestScheduler_CRUD(t *testing.T) {
	_, router, token := setupTestHandler(t)

	// Create
	w := doRequest(router, "POST", "/api/v1/schedulers", map[string]interface{}{
		"name": "Daily Check", "type": "daily",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create scheduler failed: %d", w.Code)
	}
	resp := parseResponse(w)
	data := resp.Data.(map[string]interface{})
	id := data["id"].(string)
	if data["next_run"] == "" {
		t.Error("expected next_run to be calculated")
	}

	// List
	w = doRequest(router, "GET", "/api/v1/schedulers", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("list expected 200, got %d", w.Code)
	}

	// Get
	w = doRequest(router, "GET", "/api/v1/schedulers/"+id, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("get expected 200, got %d", w.Code)
	}

	// Update
	w = doRequest(router, "PUT", "/api/v1/schedulers/"+id, map[string]interface{}{
		"name": "Updated Scheduler", "type": "monthly",
	}, token)
	if w.Code != http.StatusOK {
		t.Errorf("update expected 200, got %d", w.Code)
	}

	// Delete
	w = doRequest(router, "DELETE", "/api/v1/schedulers/"+id, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("delete expected 200, got %d", w.Code)
	}

	// Get after delete
	w = doRequest(router, "GET", "/api/v1/schedulers/"+id, nil, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
}

func TestScheduler_Create_CronType(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/schedulers", map[string]interface{}{
		"name": "Cron Job", "type": "cron", "cron_expression": "0 9 * * 1",
	}, token)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

// --- Resource CRUD Tests ---

func TestResource_CRUD(t *testing.T) {
	_, router, token := setupTestHandler(t)

	// Create
	w := doRequest(router, "POST", "/api/v1/resources", map[string]interface{}{
		"name": "Plumber", "type": "Project", "linked_items": []string{},
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create resource failed: %d", w.Code)
	}
	resp := parseResponse(w)
	data := resp.Data.(map[string]interface{})
	id := data["id"].(string)

	// List
	w = doRequest(router, "GET", "/api/v1/resources", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("list expected 200, got %d", w.Code)
	}
	resp = parseResponse(w)
	items := resp.Data.([]interface{})
	if len(items) != 1 {
		t.Errorf("expected 1 resource, got %d", len(items))
	}

	// Get
	w = doRequest(router, "GET", "/api/v1/resources/"+id, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("get expected 200, got %d", w.Code)
	}

	// Update
	w = doRequest(router, "PUT", "/api/v1/resources/"+id, map[string]interface{}{
		"name": "Electrician", "type": "Global", "linked_items": []string{"abc"},
	}, token)
	if w.Code != http.StatusOK {
		t.Errorf("update expected 200, got %d", w.Code)
	}
	resp = parseResponse(w)
	data = resp.Data.(map[string]interface{})
	if data["name"] != "Electrician" {
		t.Errorf("expected 'Electrician', got '%v'", data["name"])
	}

	// Delete
	w = doRequest(router, "DELETE", "/api/v1/resources/"+id, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("delete expected 200, got %d", w.Code)
	}

	// Get after delete
	w = doRequest(router, "GET", "/api/v1/resources/"+id, nil, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
}

func TestResource_Create_InvalidType(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/resources", map[string]interface{}{
		"name": "Bad", "type": "invalid",
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- Google OAuth Handler Tests ---

func TestGoogleLogin_NotConfigured(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "GET", "/api/v1/auth/google/login", nil, token)
	// Our handler should respond with 503 when google_client_id is empty
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for unconfigured Google OAuth, got %d", w.Code)
	}
}

func TestGoogleCallback_MissingCode(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "GET", "/api/v1/auth/google/callback", nil, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing code, got %d", w.Code)
	}
}

// --- Task Create with Valid Reminders ---

func TestTask_Create_WithReminders(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	w := doRequest(router, "POST", "/api/v1/tasks", map[string]interface{}{
		"board_id": b["id"].(string), "swimlane": "To Do", "task_type": "Feature",
		"title": "With Reminders", "priority": "High",
		"reminders": []map[string]string{
			{"time": "2026-06-01T09:00:00Z", "note": "First reminder"},
			{"time": "2026-06-15T09:00:00Z", "note": "Second reminder"},
		},
		"cost": 100.50, "estimation_minutes": 120,
	}, token)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d; body: %s", w.Code, w.Body.String())
	}
	resp := parseResponse(w)
	data := resp.Data.(map[string]interface{})
	reminders := data["reminders"].([]interface{})
	if len(reminders) != 2 {
		t.Errorf("expected 2 reminders, got %d", len(reminders))
	}
}

func TestTask_Create_InvalidBoard(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/tasks", map[string]interface{}{
		"board_id": "nonexistent", "swimlane": "To Do", "task_type": "Feature",
		"title": "Bad", "priority": "Medium",
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid board, got %d", w.Code)
	}
}

func TestTask_Create_InvalidTaskType(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	w := doRequest(router, "POST", "/api/v1/tasks", map[string]interface{}{
		"board_id": b["id"].(string), "swimlane": "To Do", "task_type": "Invalid",
		"title": "Bad", "priority": "Medium",
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid task type, got %d", w.Code)
	}
}

func TestTask_Update_TooManyReminders(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	task := createTask(t, router, token, b["id"].(string))
	tid := task["id"].(string)
	reminders := make([]map[string]string, 6)
	for i := range reminders {
		reminders[i] = map[string]string{"time": "2026-01-01T00:00:00Z", "note": "r"}
	}
	w := doRequest(router, "PUT", "/api/v1/tasks/"+tid, map[string]interface{}{
		"board_id": b["id"].(string), "swimlane": "To Do", "task_type": "Feature",
		"title": "Updated", "priority": "Medium", "reminders": reminders,
	}, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for >5 reminders on update, got %d", w.Code)
	}
}

func TestTask_Update_NotFound(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "PUT", "/api/v1/tasks/nonexistent", map[string]interface{}{
		"board_id": "x", "swimlane": "To Do", "task_type": "Feature",
		"title": "x", "priority": "Medium",
	}, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// --- Patch with various fields ---

func TestTask_PatchMultipleFields(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	task := createTask(t, router, token, b["id"].(string))
	tid := task["id"].(string)

	w := doRequest(router, "PATCH", "/api/v1/tasks/"+tid, map[string]interface{}{
		"title":               "Patched Title",
		"description":         "Patched Desc",
		"priority":            "Critical",
		"cost":                250.50,
		"estimation_minutes":  60.0,
		"actual_time_minutes": 45.0,
		"task_type":           "Bug",
	}, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	resp := parseResponse(w)
	data := resp.Data.(map[string]interface{})
	if data["title"] != "Patched Title" {
		t.Errorf("expected 'Patched Title', got '%v'", data["title"])
	}
	if data["priority"] != "Critical" {
		t.Errorf("expected 'Critical', got '%v'", data["priority"])
	}
}

// --- Scheduler with yearly type ---

func TestScheduler_YearlyType(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "POST", "/api/v1/schedulers", map[string]interface{}{
		"name": "Yearly Check", "type": "yearly",
	}, token)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	resp := parseResponse(w)
	data := resp.Data.(map[string]interface{})
	if data["next_run"] == "" {
		t.Error("expected next_run to be set")
	}
}

// --- Edge Cases ---

func TestUpdate_NotFound(t *testing.T) {
	_, router, token := setupTestHandler(t)

	w := doRequest(router, "PUT", "/api/v1/projects/nonexistent", map[string]interface{}{
		"name": "x", "type": "personal",
	}, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("project update not found: expected 404, got %d", w.Code)
	}

	w = doRequest(router, "PUT", "/api/v1/boards/nonexistent", map[string]interface{}{
		"name": "x", "project_id": "y",
	}, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("board update not found: expected 404, got %d", w.Code)
	}

	w = doRequest(router, "PUT", "/api/v1/schedulers/nonexistent", map[string]interface{}{
		"name": "x", "type": "daily",
	}, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("scheduler update not found: expected 404, got %d", w.Code)
	}

	w = doRequest(router, "PUT", "/api/v1/resources/nonexistent", map[string]interface{}{
		"name": "x", "type": "Global",
	}, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("resource update not found: expected 404, got %d", w.Code)
	}
}

func TestDelete_NotFound(t *testing.T) {
	_, router, token := setupTestHandler(t)

	for _, entity := range []string{"boards", "tasks", "schedulers", "resources"} {
		w := doRequest(router, "DELETE", "/api/v1/"+entity+"/nonexistent", nil, token)
		if w.Code != http.StatusNotFound {
			t.Errorf("%s delete not found: expected 404, got %d", entity, w.Code)
		}
	}
}

func TestPatch_NotFound(t *testing.T) {
	_, router, token := setupTestHandler(t)
	w := doRequest(router, "PATCH", "/api/v1/tasks/nonexistent", map[string]interface{}{
		"swimlane": "Done",
	}, token)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTask_DueDate(t *testing.T) {
	_, router, token := setupTestHandler(t)
	p := createProject(t, router, token)
	b := createBoard(t, router, token, p["id"].(string))
	bid := b["id"].(string)

	// 1. Create task with due_date
	dueDate := "2026-06-27T21:33:06Z"
	w := doRequest(router, "POST", "/api/v1/tasks", map[string]interface{}{
		"board_id": bid, "swimlane": "To Do", "task_type": "Feature",
		"title": "Task with due date", "priority": "Medium", "due_date": dueDate,
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d %s", w.Code, w.Body.String())
	}
	resp := parseResponse(w)
	taskData := resp.Data.(map[string]interface{})
	tid := taskData["id"].(string)
	if taskData["due_date"] != dueDate {
		t.Errorf("expected due_date '%s', got '%v'", dueDate, taskData["due_date"])
	}

	// 2. Update task with a different due_date
	newDueDate := "2026-07-27T21:33:06Z"
	w = doRequest(router, "PUT", "/api/v1/tasks/"+tid, map[string]interface{}{
		"board_id": bid, "swimlane": "To Do", "task_type": "Feature",
		"title": "Task with due date", "priority": "Medium", "due_date": newDueDate,
	}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	resp = parseResponse(w)
	taskData = resp.Data.(map[string]interface{})
	if taskData["due_date"] != newDueDate {
		t.Errorf("expected updated due_date '%s', got '%v'", newDueDate, taskData["due_date"])
	}

	// 3. Patch task to update due_date
	patchDueDate := "2026-08-27T21:33:06Z"
	w = doRequest(router, "PATCH", "/api/v1/tasks/"+tid, map[string]interface{}{
		"due_date": patchDueDate,
	}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	resp = parseResponse(w)
	taskData = resp.Data.(map[string]interface{})
	if taskData["due_date"] != patchDueDate {
		t.Errorf("expected patched due_date '%s', got '%v'", patchDueDate, taskData["due_date"])
	}

	// 4. Clear due_date via PUT
	w = doRequest(router, "PUT", "/api/v1/tasks/"+tid, map[string]interface{}{
		"board_id": bid, "swimlane": "To Do", "task_type": "Feature",
		"title": "Task with due date", "priority": "Medium", "due_date": "",
	}, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	resp = parseResponse(w)
	taskData = resp.Data.(map[string]interface{})
	if taskData["due_date"] != nil && taskData["due_date"] != "" {
		t.Errorf("expected empty due_date, got '%v'", taskData["due_date"])
	}
}
