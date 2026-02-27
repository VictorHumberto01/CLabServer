package ws

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/vitub/CLabServer/internal/ai"
	"github.com/vitub/CLabServer/internal/initializers"
	"github.com/vitub/CLabServer/internal/models"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 524288 // 512KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Hub *Hub

	conn *websocket.Conn
	send chan []byte

	UserID   string
	UserDBID uint
	Role     string
	Name     string

	mu      sync.Mutex
	ptyFile *os.File
	cmd     *exec.Cmd
}

type WSMsg struct {
	Type       string `json:"type"`
	Payload    string `json:"payload"`
	Rows       int    `json:"rows,omitempty"`
	Cols       int    `json:"cols,omitempty"`
	ExerciseID uint   `json:"exerciseId,omitempty"`
}

type MonitorMsg struct {
	Type      string `json:"type"`
	UserID    string `json:"userId"`
	UserName  string `json:"userName"`
	Payload   string `json:"payload,omitempty"`
	Timestamp string `json:"timestamp"`
}

type AnalysisPayload struct {
	Status  string `json:"status"`
	Content string `json:"content"`
}

func (c *Client) sendAIAnalysis(analysis string, status string) {
	payload := AnalysisPayload{Status: status, Content: analysis}
	payloadBytes, _ := json.Marshal(payload)

	msg := WSMsg{
		Type:    "ai_analysis",
		Payload: string(payloadBytes),
	}
	jsonBytes, _ := json.Marshal(msg)
	select {
	case c.send <- jsonBytes:
	default:
		log.Println("WS Send Buffer Full, dropping AI analysis")
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.conn.Close()

		c.mu.Lock()
		if c.ptyFile != nil {
			c.ptyFile.Close()
		}
		if c.cmd != nil && c.cmd.Process != nil {
			c.cmd.Process.Kill()
		}
		c.mu.Unlock()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg WSMsg
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		switch msg.Type {
		case "input":
			c.mu.Lock()
			if c.ptyFile != nil {
				c.ptyFile.Write([]byte(msg.Payload))
			}
			c.mu.Unlock()
		case "resize":
			c.mu.Lock()
			if c.ptyFile != nil {
				pty.Setsize(c.ptyFile, &pty.Winsize{
					Rows: uint16(msg.Rows),
					Cols: uint16(msg.Cols),
					X:    0,
					Y:    0,
				})
			}
			c.mu.Unlock()
		case "run_code":
			c.mu.Lock()
			if c.cmd != nil && c.cmd.Process != nil {
				c.cmd.Process.Kill()
			}
			c.mu.Unlock()
			go c.startCompilationAndRun(msg.Payload, msg.ExerciseID)
		case "stop":
			c.mu.Lock()
			if c.cmd != nil && c.cmd.Process != nil {
				c.cmd.Process.Kill()
				c.sendOutput("\r\n[User Interruption]: Process killed by user.\r\n")
			}
			c.mu.Unlock()
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	user, exists := c.Get("user")
	var userID string
	var userDBID uint
	var role, name string
	if exists {
		u := user.(models.User)
		userID = fmt.Sprintf("%d", u.ID)
		userDBID = u.ID
		role = u.Role
		name = u.Name
	} else {
		userID = "anon"
		role = "GUEST"
		name = "Anonymous"
	}

	client := &Client{
		Hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		UserID:   userID,
		UserDBID: userDBID,
		Role:     role,
		Name:     name,
	}
	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) startCompilationAndRun(code string, exerciseID uint) {
	log.Printf("WS: Starting Run/Submission. UserID: %s (DBID: %d), Name: %s, ExerciseID: %d", c.UserID, c.UserDBID, c.Name, exerciseID)
	c.broadcastMonitor("compile_start", "Starting compilation...")

	tmpDir, err := os.MkdirTemp("", "cws")
	if err != nil {
		c.sendOutput("Error creating temp dir: " + err.Error())
		return
	}
	defer os.RemoveAll(tmpDir)

	srcPath := tmpDir + "/program.c"
	binPath := tmpDir + "/program"
	os.WriteFile(srcPath, []byte(code), 0644)

	var isExam bool
	if exerciseID > 0 {
		var exercise models.Exercise
		if err := initializers.DB.Preload("Topic").First(&exercise, exerciseID).Error; err == nil {
			isExam = exercise.Topic != nil && exercise.Topic.IsExam
		}
	}

	compileCmd := exec.Command("gcc", srcPath, "-o", binPath, "-Wall")
	out, err := compileCmd.CombinedOutput()
	if err != nil {
		errorOutput := string(out)

		c.sendOutput("Compilation Error:\r\n" + errorOutput)

		var analysis string
		var aiErr error

		if !isExam {
			analysis, aiErr = ai.GetErrorAnalysis(code, errorOutput)
			if aiErr == nil {
				c.sendAIAnalysis(analysis, "error")
				c.sendOutput("\r\n[AI]: Compilation analysis sent to side panel.\r\n")
			}
		} else {
			c.sendOutput("\r\n[MODO PROVA]: Detalhes do erro suprimidos. Verifique sua sintaxe.\r\n")
		}

		c.broadcastMonitor("compile_end", "Compilation failed")

		statusMsg := WSMsg{
			Type:    "status",
			Payload: "stopped",
		}
		if statusBytes, err := json.Marshal(statusMsg); err == nil {
			c.sendOutput(string(statusBytes))
		}

		if c.UserDBID != 0 {
			history := models.History{
				UserID:    c.UserDBID,
				Code:      code,
				Error:     errorOutput,
				IsSuccess: false,
				Score:     0,
			}
			if exerciseID > 0 {
				exID := exerciseID
				history.ExerciseID = &exID
			}

			if isExam {
				gradingRes, err := ai.GetExamErrorAnalysis(code, errorOutput)
				if err == nil {
					history.TeacherGrading = gradingRes.Feedback
					history.Score = gradingRes.Score
				}
			} else {
				history.AIAnalysis = analysis
			}

			initializers.DB.Create(&history)
		}

		return
	}
	c.sendOutput("Compilation successful.\r\nRunning...\r\n")

	runCmd := exec.Command(binPath)
	ptyFile, err := pty.Start(runCmd)
	if err != nil {
		c.sendOutput("Error starting PTY: " + err.Error())
		return
	}

	c.mu.Lock()
	c.ptyFile = ptyFile
	c.cmd = runCmd
	c.mu.Unlock()
	buf := make([]byte, 1024)
	var fullOutput []byte
	var aiAnalysisStored string
	for {
		n, err := ptyFile.Read(buf)
		if n > 0 {
			output := buf[:n]
			fullOutput = append(fullOutput, output...)
			c.sendOutput(string(output))
			c.broadcastMonitor("output_chunk", string(output))
		}
		if err != nil {
			if err != io.EOF {
			}
			break
		}
	}

	err = runCmd.Wait()

	exitMsg := "\r\nProgram exited."
	isSuccess := false

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitMsg = fmt.Sprintf("\r\nProgram exited with code %d", exitErr.ExitCode())

			if err.Error() == "signal: killed" {
				c.sendOutput(fmt.Sprintf("\r\n[Runtime Error]: %v", err))
			} else {
				c.sendOutput(fmt.Sprintf("\r\n[Runtime Error]: %v\r\nAnalyzing...", err))

				analysis, aiErr := ai.GetErrorAnalysis(code, string(fullOutput))
				if aiErr != nil {
					c.sendOutput("\r\nAI Analysis failed: " + aiErr.Error())
				} else {
					aiAnalysisStored = analysis
					c.sendAIAnalysis(analysis, "error")
					c.sendOutput("\r\n[AI]: Analysis sent to side panel.\r\n")
				}
			}
		} else {
			exitMsg = fmt.Sprintf("\r\nProgram exited with error: %v", err)
		}
	} else {
		isSuccess = true

		if exerciseID == 0 {
			// Normal code run - perform AI analysis
			c.sendOutput("\r\nAnalyzing...")
			analysis, aiErr := ai.GetAIAnalysis(code, string(fullOutput))
			if aiErr != nil {
				c.sendOutput("\r\nAI Analysis failed: " + aiErr.Error())
			} else {
				aiAnalysisStored = analysis
				c.sendAIAnalysis(analysis, "success")
			}
		} else if exerciseID > 0 {
			c.sendOutput("\r\nAvaliando Exercício...")
			c.sendOutput("\r\nAvaliando Exercício...")
			var exercise models.Exercise
			if dbErr := initializers.DB.Preload("Topic").First(&exercise, exerciseID).Error; dbErr != nil {
				c.sendOutput("\r\nFalha ao buscar detalhes do exercício: " + dbErr.Error())
			} else {

				isExam := exercise.Topic != nil && exercise.Topic.IsExam

				if isExam {
					c.sendOutput("\r\n[MODO PROVA]: Submissão recebida.")
					c.sendOutput("\r\nO professor receberá sua resposta para correção.")

					grading, aiErr := ai.GetExamGradingAnalysis(code, string(fullOutput), exercise.ExpectedOutput, exercise.ExamMaxNote)
					if aiErr != nil {
						log.Printf("Exam grading failed: %v", aiErr)
						aiAnalysisStored = "Erro na correção automática: " + aiErr.Error()
					} else {
						gradingJson, _ := json.Marshal(grading)
						aiAnalysisStored = string(gradingJson)

					}
					isSuccess = true

				} else {
					analysis, aiErr := ai.GetAIAnalysis(code, string(fullOutput))
					if aiErr != nil {
						c.sendOutput("\r\nAI Analysis failed: " + aiErr.Error())
					} else {
						aiAnalysisStored = analysis
						c.sendAIAnalysis(analysis, "success")
					}
					isSuccess = true
				}
			}
		}
	}

	c.sendOutput(exitMsg)
	c.broadcastMonitor("compile_end", exitMsg)

	statusMsg := WSMsg{
		Type:    "status",
		Payload: "stopped",
	}
	if statusBytes, err := json.Marshal(statusMsg); err == nil {
		c.sendOutput(string(statusBytes))
	}

	if c.UserDBID != 0 {
		var teacherGradingData string
		var scoreVal float64

		if len(aiAnalysisStored) > 0 && aiAnalysisStored[0] == '{' {
			var gradingRes ai.ExamGradingResult
			if err := json.Unmarshal([]byte(aiAnalysisStored), &gradingRes); err == nil {
				teacherGradingData = gradingRes.Feedback
				scoreVal = gradingRes.Score
				aiAnalysisStored = ""
			}
		}

		history := models.History{
			UserID:         c.UserDBID,
			Code:           code,
			Output:         string(fullOutput),
			AIAnalysis:     aiAnalysisStored, // Empty for exams if we cleared it
			TeacherGrading: teacherGradingData,
			Score:          scoreVal,
			IsSuccess:      isSuccess,
		}
		if exerciseID > 0 {
			exID := exerciseID
			history.ExerciseID = &exID
		}

		if err := initializers.DB.Create(&history).Error; err != nil {
			log.Printf("Failed to save history for user %d: %v", c.UserDBID, err)
		} else {
			log.Printf("Saved history for user %d (Exercise: %d, Success: %v)", c.UserDBID, exerciseID, isSuccess)
		}
	}

	c.mu.Lock()
	c.ptyFile = nil
	c.cmd = nil
	c.mu.Unlock()
}

func (c *Client) sendOutput(text string) {
	select {
	case c.send <- []byte(text):
	default:

		log.Println("WS Send Buffer Full, dropping output")
	}
}

func (c *Client) broadcastMonitor(msgType, payload string) {
	msg := MonitorMsg{
		Type:      msgType,
		UserID:    c.UserID,
		UserName:  c.Name,
		Payload:   payload,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	jsonBytes, _ := json.Marshal(msg)
	c.Hub.BroadcastToMonitors(jsonBytes)
}
