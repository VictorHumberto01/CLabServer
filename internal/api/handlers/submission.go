package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"clab-server/internal/compiler/executor"
	"clab-server/internal/database/models"
)

type SubmissionHandler struct {
	db       *gorm.DB
	executor *executor.Executor
}

func NewSubmissionHandler(db *gorm.DB, executor *executor.Executor) *SubmissionHandler {
	return &SubmissionHandler{
		db:       db,
		executor: executor,
	}
}

type SubmitRequest struct {
	Code string `json:"code"`
}

func (h *SubmissionHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	taskID, err := strconv.ParseUint(vars["task_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Get task and its test cases
	var task models.Task
	if err := h.db.Preload("TestCases").First(&task, taskID).Error; err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Create submission
	submission := models.Submission{
		TaskID: uint(taskID),
		UserID: r.Context().Value("user_id").(uint),
		Code:   req.Code,
		Status: "pending",
	}

	if err := h.db.Create(&submission).Error; err != nil {
		http.Error(w, "Failed to create submission", http.StatusInternalServerError)
		return
	}

	// Run test cases
	var feedback string
	var score float64
	totalTests := len(task.TestCases)
	passedTests := 0

	for _, testCase := range task.TestCases {
		result, err := h.executor.CompileAndRun(req.Code, testCase.Input)
		if err != nil {
			feedback += "Error: " + err.Error() + "\n"
			continue
		}

		if result.ExitCode != 0 {
			feedback += "Compilation error: " + result.Error + "\n"
			continue
		}

		if result.Output == testCase.ExpectedOutput {
			passedTests++
			feedback += "Test passed\n"
		} else {
			feedback += "Test failed\n"
			feedback += "Expected: " + testCase.ExpectedOutput + "\n"
			feedback += "Got: " + result.Output + "\n"
		}
	}

	// Calculate score
	score = float64(passedTests) / float64(totalTests) * 100

	// Update submission
	submission.Status = "completed"
	submission.Feedback = feedback
	submission.Score = score

	if err := h.db.Save(&submission).Error; err != nil {
		http.Error(w, "Failed to update submission", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(submission)
}

func (h *SubmissionHandler) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.ParseUint(vars["task_id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uint)
	role := r.Context().Value("role").(string)

	var submissions []models.Submission
	query := h.db.Where("task_id = ?", taskID)

	// Students can only see their own submissions
	if role == string(models.RoleStudent) {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&submissions).Error; err != nil {
		http.Error(w, "Failed to list submissions", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(submissions)
}
