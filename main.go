package main

import (
	"bufio"
	"bytes"
	"encoding/json"
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

type CodeAnalysis struct {
	Elements    []CodeElement `json:"elements"`
	Suggestions []string      `json:"suggestions"`
	AIAnalysis  string        `json:"aiAnalysis"`
}

type CodeElement struct {
	Element     string `json:"element"`
	Description string `json:"description"`
}

type CompileResponse struct {
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	Analysis string `json:"analysis"`
}

// Add this function to get AI analysis for successful compilation
func getAIAnalysis(code string) (string, error) {
	prompt := fmt.Sprintf(`Analise este código C e forneça uma explicação detalhada em português.
Você é um professor experiente explicando o código para um aluno.

Código para análise:
%s

Formate sua resposta exatamente assim:
===Analysis===
# Análise Detalhada do Código

Este programa foi criado para [explicar o propósito]. Vamos analisar cada parte:

## Estrutura Básica
[explicar a estrutura do código]

## Bibliotecas e Funções
[explicar as bibliotecas e funções usadas]

## Funcionamento
[explicar como o código funciona]

## Sugestões de Melhoria
[listar sugestões de melhoria]

## Dicas de Aprendizado
[incluir dicas educacionais]`, code)

	return callOllamaAPI(prompt)
}

// Add this function to get AI analysis for compilation errors
func getErrorAnalysis(code string, errorMessage string) (string, error) {
	prompt := fmt.Sprintf(`Analise este código C que teve erro de compilação e explique detalhadamente o problema em português.
Você é um professor experiente ajudando um aluno a entender e corrigir erros.

Código com erro:
%s

Mensagem de erro do compilador:
%s

Formate sua resposta exatamente assim:
===Analysis===
# Análise do Erro de Compilação

## 🚫 Erro Encontrado
[explicar claramente qual foi o erro]

## 🔍 Causa do Problema
[explicar por que o erro aconteceu]

## 📚 Conceitos Importantes
[explicar os conceitos de C que o usuário precisa entender]

## ✅ Como Corrigir
[mostrar como corrigir o erro com exemplos]

## 💡 Dicas para Evitar
[dar dicas para evitar erros similares no futuro]

## 📖 Exemplo Correto
[mostrar um exemplo de código corrigido se possível]`, code, errorMessage)

	return callOllamaAPI(prompt)
}

// Helper function to call Ollama API
func callOllamaAPI(prompt string) (string, error) {
	payload := map[string]interface{}{
		"model":       "phi3",
		"system":      "Você é um professor experiente de programação C, explicando conceitos para um aluno de forma didática e clara.",
		"prompt":      prompt,
		"stream":      false,
		"temperature": 0.5,
		"top_p":       0.9,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	resp, err := http.Post("http://localhost:11434/api/generate",
		"application/json",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error calling Ollama API: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	response, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format from Ollama")
	}

	return response, nil
}

func compileAndRun(code string) CompileResponse {
	log.Println("Starting compilation process")
	tmpDir, err := os.MkdirTemp("", "ccompile")
	if err != nil {
		log.Printf("Failed to create temp directory: %v", err)
		return CompileResponse{Error: "failed to create temp dir"}
	}
	defer os.RemoveAll(tmpDir)
	log.Printf("Created temporary directory: %s", tmpDir)

	srcPath := filepath.Join(tmpDir, "program.c")
	binPath := filepath.Join(tmpDir, "program")

	// Escreve o código
	if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
		log.Printf("Failed to write source file: %v", err)
		return CompileResponse{Error: "failed to write source"}
	}
	log.Printf("Source code written to: %s", srcPath)

	// Compila
	compileCmd := exec.Command("gcc", srcPath, "-o", binPath)
	compileCmd.Dir = tmpDir
	log.Printf("Compiling with command: gcc %s -o %s", srcPath, binPath)
	compileOut, err := compileCmd.CombinedOutput()

	// Se houve erro de compilação, retorna com análise do erro
	if err != nil {
		log.Printf("Compilation failed: %v\nOutput: %s", err, string(compileOut))

		// Gera análise do erro
		errorAnalysis, analysisErr := getErrorAnalysis(code, string(compileOut))
		if analysisErr != nil {
			log.Printf("Error analysis failed: %v", analysisErr)
			errorAnalysis = "===Analysis===\n# Análise do Erro\n\nDesculpe, não foi possível gerar a análise detalhada do erro neste momento. Por favor, verifique a mensagem de erro do compilador acima."
		} else {
			log.Printf("Error analysis generated:\n%s", errorAnalysis)
		}

		return CompileResponse{
			Error:    "Compilation error: " + string(compileOut),
			Analysis: errorAnalysis,
		}
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

			// Mesmo sem execução, gera análise do código compilado com sucesso
			analysis, err := getAIAnalysis(code)
			if err != nil {
				log.Printf("AI analysis error: %v", err)
				analysis = "===Analysis===\n# Análise do Código\n\nDesculpe, não foi possível gerar a análise neste momento. Por favor, tente novamente."
			} else {
				log.Printf("AI Analysis generated:\n%s", analysis)
			}

			return CompileResponse{
				Output:   "Compilation successful, but execution was skipped because firejail is not available and unsecure mode is disabled.",
				Error:    "execution disabled: firejail not available and unsecure mode not allowed",
				Analysis: analysis,
			}
		}
	}

	execCmd.Dir = tmpDir

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
		// Sempre gera análise, independente se a execução foi bem-sucedida ou não
		analysis, analysisErr := getAIAnalysis(code)
		if analysisErr != nil {
			log.Printf("AI analysis error: %v", analysisErr)
			analysis = "===Analysis===\n# Análise do Código\n\nDesculpe, não foi possível gerar a análise neste momento. Por favor, tente novamente."
		} else {
			log.Printf("AI Analysis generated:\n%s", analysis)
		}

		if err != nil {
			log.Printf("Program execution failed: %v", err)
			return CompileResponse{
				Output:   output.String(),
				Error:    "execution failed: " + err.Error(),
				Analysis: analysis,
			}
		} else {
			log.Println("Program execution completed successfully")
			return CompileResponse{
				Output:   output.String(),
				Analysis: analysis,
			}
		}
	case <-time.After(3 * time.Second):
		log.Println("Program execution timed out")
		if execCmd.Process != nil {
			execCmd.Process.Kill()
		}

		// Gera análise mesmo em caso de timeout
		analysis, err := getAIAnalysis(code)
		if err != nil {
			log.Printf("AI analysis error: %v", err)
			analysis = "===Analysis===\n# Análise do Código\n\nDesculpe, não foi possível gerar a análise neste momento. Por favor, tente novamente."
		} else {
			log.Printf("AI Analysis generated:\n%s", analysis)
		}

		return CompileResponse{
			Error:    "execution timeout",
			Analysis: analysis,
		}
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
		result := compileAndRun(req.Code)

		c.JSON(http.StatusOK, result)
	})

	router.Run(":8080")
}
