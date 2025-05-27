package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Global variables to track unsecure mode settings
var (
	unsecureModePrompted bool = false
	unsecureModeAllowed  bool = false
	unsecureModeMutex    sync.Mutex
)

// Check if a command is available in the system PATH
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// promptForUnsecureMode asks the user once if they want to allow unsecure execution
// when firejail is not available. The choice persists for the lifetime of the server.
func promptForUnsecureMode() bool {
	unsecureModeMutex.Lock()
	defer unsecureModeMutex.Unlock()

	// If we've already prompted, return the saved choice
	if unsecureModePrompted {
		return unsecureModeAllowed
	}

	fmt.Println("\n⚠️ WARNING: Firejail is not available on this system!")
	fmt.Println("Running code without sandboxing is a security risk.")
	fmt.Print("Do you want to allow execution in unsecure mode? (y/n): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	// Mark as prompted so we don't ask again
	unsecureModePrompted = true

	// Set the user's choice
	if input == "y" || input == "yes" {
		fmt.Println("Are you sure? ")
		fmt.Print("Type 'yes' to confirm: ")
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "yes" {
			fmt.Println("Unsecure mode enabled. All code will run without sandboxing.")
			fmt.Println("Please note that this may compromise your system security and cause harm to data.")
			unsecureModeAllowed = true
			return true
		}
	}

	fmt.Println("Unsecure mode rejected. Compilation will continue, but execution will be disabled.")
	unsecureModeAllowed = false
	return false
}

type CompileRequest struct {
	Code  string `json:"code"`
	Input string `json:"input"`
}

type CompileResult struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func compileAndRun(code string, input string, resultChan chan CompileResult) {
	log.Println("Starting compilation process")
	tmpDir, err := os.MkdirTemp("", "ccompile")
	if err != nil {
		log.Printf("Failed to create temp directory: %v", err)
		resultChan <- CompileResult{Error: "failed to create temp dir"}
		return
	}
	defer os.RemoveAll(tmpDir)
	log.Printf("Created temporary directory: %s", tmpDir)

	srcPath := filepath.Join(tmpDir, "program.c")
	binPath := filepath.Join(tmpDir, "program")

	// Escreve o código
	if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
		log.Printf("Failed to write source file: %v", err)
		resultChan <- CompileResult{Error: "failed to write source"}
		return
	}
	log.Printf("Source code written to: %s", srcPath)

	// Compila
	compileCmd := exec.Command("gcc", srcPath, "-o", binPath)
	compileCmd.Dir = tmpDir
	log.Printf("Compiling with command: gcc %s -o %s", srcPath, binPath)
	compileOut, err := compileCmd.CombinedOutput()
	if err != nil {
		log.Printf("Compilation failed: %v\nOutput: %s", err, string(compileOut))
		resultChan <- CompileResult{Error: "Compilation error: " + string(compileOut)}
		return
	}
	log.Println("Compilation successful")

	// Check if firejail is available
	var execCmd *exec.Cmd
	if isCommandAvailable("firejail") {
		log.Println("Firejail is available, using it for sandboxed execution")
		execCmd = exec.Command("firejail", "--quiet", "--net=none", "--private="+tmpDir, "./program")
	} else {
		// Check if unsecure mode is allowed
		if promptForUnsecureMode() {
			log.Println("Warning: Running in unsecure mode without firejail")
			execCmd = exec.Command("./program")
		} else {
			log.Println("Execution skipped: unsecure mode not allowed by user")
			resultChan <- CompileResult{
				Output: "Compilation successful, but execution was skipped because firejail is not available and unsecure mode is disabled.",
				Error:  "execution disabled: firejail not available and unsecure mode not allowed",
			}
			return
		}
	}

	execCmd.Dir = tmpDir
	if input != "" {
		execCmd.Stdin = bytes.NewBufferString(input)
		log.Printf("Providing input to program: %s", input)
	}

	var output bytes.Buffer
	execCmd.Stdout = &output
	execCmd.Stderr = &output

	log.Println("Starting program execution")
	done := make(chan error, 1)
	go func() {
		done <- execCmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Printf("Program execution failed: %v", err)
			resultChan <- CompileResult{Output: output.String(), Error: "execution failed: " + err.Error()}
		} else {
			log.Println("Program execution completed successfully")
			resultChan <- CompileResult{Output: output.String()}
		}
	case <-time.After(3 * time.Second):
		log.Println("Program execution timed out")
		if execCmd.Process != nil {
			execCmd.Process.Kill()
		}
		resultChan <- CompileResult{Error: "execution timeout"}
	}
}

func main() {
	router := gin.Default()

	// Configure CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	router.POST("/compile", func(c *gin.Context) {
		var req CompileRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		log.Printf("Received compile request with %d bytes of code", len(req.Code))
		resultChan := make(chan CompileResult)
		go compileAndRun(req.Code, req.Input, resultChan)

		result := <-resultChan
		log.Printf("Compilation result: output=%d bytes, error=%v",
			len(result.Output), result.Error != "")
		c.JSON(http.StatusOK, result)
	})

	router.Run(":8080")
}
