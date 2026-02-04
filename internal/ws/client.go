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

	UserID string
	Role   string
	Name   string

	mu      sync.Mutex
	ptyFile *os.File
	cmd     *exec.Cmd
}

type WSMsg struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Rows    int    `json:"rows,omitempty"`
	Cols    int    `json:"cols,omitempty"`
}

type MonitorMsg struct {
	Type      string `json:"type"`
	UserID    string `json:"userId"`
	UserName  string `json:"userName"`
	Payload   string `json:"payload,omitempty"`
	Timestamp string `json:"timestamp"`
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
			go c.startCompilationAndRun(msg.Payload)
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

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

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
	var userID, role, name string
	if exists {
		u := user.(models.User)
		userID = string(rune(u.ID))
		role = u.Role
		name = u.Name
	} else {
		userID = "anon"
		role = "GUEST"
		name = "Anonymous"
	}

	client := &Client{
		Hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		UserID: userID,
		Role:   role,
		Name:   name,
	}
	client.Hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) startCompilationAndRun(code string) {
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

	compileCmd := exec.Command("gcc", srcPath, "-o", binPath, "-Wall")
	out, err := compileCmd.CombinedOutput()
	if err != nil {
		errorOutput := string(out)

		c.sendOutput("Compilation Error:\r\n" + errorOutput)
		c.broadcastMonitor("compile_end", "Compilation failed")
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
					msg := WSMsg{
						Type:    "ai_analysis",
						Payload: analysis,
					}
					jsonBytes, _ := json.Marshal(msg)
					select {
					case c.send <- jsonBytes:
					default:
						log.Println("WS Send Buffer Full, dropping AI analysis")
					}

					c.sendOutput("\r\n[AI]: Analysis sent to side panel.\r\n")
				}
			}
		} else {
			exitMsg = fmt.Sprintf("\r\nProgram exited with error: %v", err)
		}
	} else {
		c.sendOutput("\r\nAnalyzing...")
		analysis, aiErr := ai.GetAIAnalysis(code, string(fullOutput))
		if aiErr != nil {
			c.sendOutput("\r\nAI Analysis failed: " + aiErr.Error())
		} else {
			msg := WSMsg{
				Type:    "ai_analysis",
				Payload: analysis,
			}
			jsonBytes, _ := json.Marshal(msg)
			select {
			case c.send <- jsonBytes:
			default:
				log.Println("WS Send Buffer Full, dropping AI analysis")
			}
		}
	}

	c.sendOutput(exitMsg)
	c.broadcastMonitor("compile_end", exitMsg)

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
