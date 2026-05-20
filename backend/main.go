// Package main is the entry point for the MyKanban backend server.
//
// @title           MyKanban Backend API
// @version         1.0
// @description     A minimalistic, secure REST API for personal and professional task tracking.
// @termsOfService  http://swagger.io/terms/
//
// @contact.name   MyKanban Support
// @contact.email  admin@mykanban.local
//
// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT
//
// @host      localhost:8080
// @BasePath  /api
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token with the `Bearer ` prefix, e.g. "Bearer eyJhbGci..."
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "mykanban-backend/docs"
	"mykanban-backend/handlers"
	"mykanban-backend/middleware"
	"mykanban-backend/models"
	"mykanban-backend/storage"
)

func main() {
	// Load configuration
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Unable to get Working Direcotry : ", err.Error())
	}
	fmt.Println("***: pwd=" + pwd + " :***")
	configPath := "backend/config.json"
	cfgStore := storage.NewConfigStore(configPath, nil)
	var cfg models.Config
	if err := cfgStore.Load(&cfg); err != nil {
		log.Fatalf("Failed to load config.json: %v", err)
	}

	// Set up logging
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// Initialize storage
	storageDir := cfg.StorageDir
	if storageDir == "" {
		storageDir = "./storage"
	}

	projectStore, err := storage.NewStore[models.Project](storageDir, "projects.json", nil)
	if err != nil {
		log.Fatalf("Failed to init project store: %v", err)
	}
	boardStore, err := storage.NewStore[models.Board](storageDir, "boards.json", nil)
	if err != nil {
		log.Fatalf("Failed to init board store: %v", err)
	}
	taskStore, err := storage.NewStore[models.Task](storageDir, "tasks.json", nil)
	if err != nil {
		log.Fatalf("Failed to init task store: %v", err)
	}
	schedulerStore, err := storage.NewStore[models.Scheduler](storageDir, "schedulers.json", nil)
	if err != nil {
		log.Fatalf("Failed to init scheduler store: %v", err)
	}
	resourceStore, err := storage.NewStore[models.Resource](storageDir, "resources.json", nil)
	if err != nil {
		log.Fatalf("Failed to init resource store: %v", err)
	}

	h := &handlers.Handler{
		Projects:   projectStore,
		Boards:     boardStore,
		Tasks:      taskStore,
		Schedulers: schedulerStore,
		Resources:  resourceStore,
		Config:     cfgStore,
		AppConfig:  &cfg,
	}

	// Set up Gin router
	router := SetupRouter(h, &cfg)

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		log.Printf("MyKanban server starting on port %d", cfg.Port)
		fmt.Printf("MyKanban server starting on port %d\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}

// parseAllowedOrigins splits a comma-separated string of origins into a slice,
// trimming whitespace from each entry. Returns a default if the input is empty.
func parseAllowedOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"http://localhost:3000"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		return []string{"http://localhost:3000"}
	}
	return origins
}

// SetupRouter configures all routes and middleware. Exported for testing.
func SetupRouter(h *handlers.Handler, cfg *models.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.Recovery())

	// CORS middleware — reads allowed origins from config.json
	allowedOrigins := parseAllowedOrigins(cfg.AllowedOrigins)
	log.Printf("CORS allowed origins: %v", allowedOrigins)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	router.GET("/api/health", handlers.Health)

	// Auth routes (public)
	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/login", h.Login)
		authGroup.GET("/google/login", h.GoogleLogin)
		authGroup.GET("/google/callback", h.GoogleCallback)
	}

	// Protected API routes
	api := router.Group("/api/v1")
	api.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Auth - change password
		api.POST("/auth/change-password", h.ChangePassword)

		// Projects
		api.POST("/projects", h.CreateProject)
		api.GET("/projects", h.ListProjects)
		api.GET("/projects/:id", h.GetProject)
		api.PUT("/projects/:id", h.UpdateProject)
		api.DELETE("/projects/:id", h.DeleteProject)

		// Boards
		api.POST("/boards", h.CreateBoard)
		api.GET("/boards", h.ListBoards)
		api.GET("/boards/:id", h.GetBoard)
		api.PUT("/boards/:id", h.UpdateBoard)
		api.DELETE("/boards/:id", h.DeleteBoard)

		// Tasks
		api.POST("/tasks", h.CreateTask)
		api.GET("/tasks", h.ListTasks)
		api.GET("/tasks/:id", h.GetTask)
		api.PUT("/tasks/:id", h.UpdateTask)
		api.PATCH("/tasks/:id", h.PatchTask)
		api.DELETE("/tasks/:id", h.DeleteTask)

		// Schedulers
		api.POST("/schedulers", h.CreateScheduler)
		api.GET("/schedulers", h.ListSchedulers)
		api.GET("/schedulers/:id", h.GetScheduler)
		api.PUT("/schedulers/:id", h.UpdateScheduler)
		api.DELETE("/schedulers/:id", h.DeleteScheduler)

		// Resources
		api.POST("/resources", h.CreateResource)
		api.GET("/resources", h.ListResources)
		api.GET("/resources/:id", h.GetResource)
		api.PUT("/resources/:id", h.UpdateResource)
		api.DELETE("/resources/:id", h.DeleteResource)
	}

	return router
}
