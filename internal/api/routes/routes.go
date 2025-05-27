package routes

import (
	"net/http"
	"time"

	"clab-server/internal/api/handlers"
	"clab-server/internal/api/middleware"
	"clab-server/internal/compiler/executor"
	"clab-server/internal/config"
	"clab-server/internal/database/models"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, authMiddleware *middleware.AuthMiddleware, cfg *config.Config) *mux.Router {
	router := mux.NewRouter()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, authMiddleware)
	roomHandler := handlers.NewRoomHandler(db)
	exec := executor.NewExecutor(cfg.CompilerPath, cfg.MaxMemoryUsage, 5*time.Second)
	submissionHandler := handlers.NewSubmissionHandler(db, exec)

	// Public routes
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")

	// Protected routes
	api := router.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware.Authenticate)

	// Admin routes
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(authMiddleware.RequireRole(string(models.RoleAdmin)))
	admin.HandleFunc("/teachers", authHandler.Register).Methods("POST")
	admin.HandleFunc("/rooms", roomHandler.ListRooms).Methods("GET")

	// Teacher routes
	teacher := api.PathPrefix("/teacher").Subrouter()
	teacher.Use(authMiddleware.RequireRole(string(models.RoleTeacher)))
	teacher.HandleFunc("/rooms", roomHandler.CreateRoom).Methods("POST")
	teacher.HandleFunc("/rooms", roomHandler.ListRooms).Methods("GET")
	teacher.HandleFunc("/rooms/{room_id}/tasks", roomHandler.CreateTask).Methods("POST")
	teacher.HandleFunc("/rooms/{room_id}/tasks", roomHandler.ListTasks).Methods("GET")
	teacher.HandleFunc("/tasks/{task_id}/submissions", submissionHandler.ListSubmissions).Methods("GET")

	// Student routes
	student := api.PathPrefix("/student").Subrouter()
	student.Use(authMiddleware.RequireRole(string(models.RoleStudent)))
	student.HandleFunc("/rooms/join", roomHandler.JoinRoom).Methods("POST")
	student.HandleFunc("/rooms", roomHandler.ListRooms).Methods("GET")
	student.HandleFunc("/rooms/{room_id}/tasks", roomHandler.ListTasks).Methods("GET")
	student.HandleFunc("/tasks/{task_id}/submit", submissionHandler.Submit).Methods("POST")
	student.HandleFunc("/tasks/{task_id}/submissions", submissionHandler.ListSubmissions).Methods("GET")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	return router
}
