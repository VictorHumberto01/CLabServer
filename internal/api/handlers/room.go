package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"clab-server/internal/database/models"
)

type RoomHandler struct {
	db *gorm.DB
}

func NewRoomHandler(db *gorm.DB) *RoomHandler {
	return &RoomHandler{db: db}
}

type CreateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	TestCases   []struct {
		Input          string `json:"input"`
		ExpectedOutput string `json:"expected_output"`
	} `json:"test_cases"`
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get teacher ID from context
	teacherID := r.Context().Value("user_id").(uint)

	room := models.Room{
		Name:        req.Name,
		Description: req.Description,
		TeacherID:   teacherID,
		Code:        generateRoomCode(), // Implement this function
	}

	if err := h.db.Create(&room).Error; err != nil {
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(room)
}

func (h *RoomHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	roomID, err := strconv.ParseUint(vars["room_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// Verify room ownership
	var room models.Room
	if err := h.db.First(&room, roomID).Error; err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	teacherID := r.Context().Value("user_id").(uint)
	if room.TeacherID != teacherID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	task := models.Task{
		RoomID:      uint(roomID),
		Title:       req.Title,
		Description: req.Description,
	}

	// Create test cases
	for _, tc := range req.TestCases {
		task.TestCases = append(task.TestCases, models.TestCase{
			Input:          tc.Input,
			ExpectedOutput: tc.ExpectedOutput,
		})
	}

	if err := h.db.Create(&task).Error; err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(task)
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Room code required", http.StatusBadRequest)
		return
	}

	var room models.Room
	if err := h.db.Where("code = ?", code).First(&room).Error; err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	studentID := r.Context().Value("user_id").(uint)

	// Add student to room
	if err := h.db.Model(&room).Association("Students").Append(&models.User{ID: studentID}); err != nil {
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(room)
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	role := r.Context().Value("role").(string)

	var rooms []models.Room
	query := h.db

	if role == string(models.RoleTeacher) {
		query = query.Where("teacher_id = ?", userID)
	} else if role == string(models.RoleStudent) {
		query = query.Joins("JOIN room_students ON room_students.room_id = rooms.id").
			Where("room_students.user_id = ?", userID)
	}

	if err := query.Find(&rooms).Error; err != nil {
		http.Error(w, "Failed to list rooms", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rooms)
}

func (h *RoomHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := strconv.ParseUint(vars["room_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	var tasks []models.Task
	if err := h.db.Where("room_id = ?", roomID).Find(&tasks).Error; err != nil {
		http.Error(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tasks)
}

// Helper function to generate a unique room code
func generateRoomCode() string {
	// Implement a function to generate a unique room code
	// This could be a random string, sequential number, etc.
	return "ROOM-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
