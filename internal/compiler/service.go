package compiler

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/vitub/CLabServer/internal/ai"
	"github.com/vitub/CLabServer/internal/models"
	"github.com/vitub/CLabServer/internal/security"
)

var securityManager *security.SecurityManager

func init() {
	securityManager = security.NewSecurityManager()
}

func CompileAndRun(req models.CompileRequest) models.CompileResponse {
	log.Println("Starting compilation process")

	tmpDir, err := os.MkdirTemp("", "ccompile")
	if err != nil {
		log.Printf("Failed to create temp directory: %v", err)
		return models.CompileResponse{Error: "failed to create temp dir"}
	}
	defer os.RemoveAll(tmpDir)
	log.Printf("Created temporary directory: %s", tmpDir)

	srcPath := filepath.Join(tmpDir, "program.c")
	binPath := filepath.Join(tmpDir, "program")

	if err := os.WriteFile(srcPath, []byte(req.Code), 0644); err != nil {
		log.Printf("Failed to write source file: %v", err)
		return models.CompileResponse{Error: "failed to write source"}
	}
	log.Printf("Source code written to: %s", srcPath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	compileCmd := exec.CommandContext(ctx, "gcc", srcPath, "-o", binPath, "-Wall", "-Wextra")
	compileCmd.Dir = tmpDir
	log.Printf("Compiling with command: gcc %s -o %s -Wall -Wextra", srcPath, binPath)
	compileOut, err := compileCmd.CombinedOutput()

	if err != nil {
		log.Printf("Compilation failed: %v\nOutput: %s", err, string(compileOut))

		errorAnalysis, analysisErr := ai.GetErrorAnalysis(req.Code, string(compileOut))
		if analysisErr != nil {
			log.Printf("Error analysis failed: %v", analysisErr)
			errorAnalysis = "===Analysis===\n# Análise do Erro\n\nDesculpe, não foi possível gerar a análise detalhada do erro neste momento. Por favor, verifique a mensagem de erro do compilador acima."
		}

		return models.CompileResponse{
			Error:    string(compileOut),
			Analysis: errorAnalysis,
		}
	}

	if err := securityManager.ValidateExecutable(binPath); err != nil {
		log.Printf("Executable validation failed: %v", err)
		return models.CompileResponse{
			Error: "Executable validation failed: " + err.Error(),
		}
	}

	timeout := 10 * time.Second
	if req.TimeoutSecs > 0 && req.TimeoutSecs <= 30 {
		timeout = time.Duration(req.TimeoutSecs) * time.Second
	}

	runCtx, runCancel := context.WithTimeout(context.Background(), timeout)
	defer runCancel()

	var runCmd *exec.Cmd

	var cmdBuilder strings.Builder
	cmdBuilder.WriteString("ulimit -v 131072; ")

	if security.IsCommandAvailable("firejail") {
		cmdBuilder.WriteString("firejail --quiet --noprofile --seccomp --nonetwork --private-tmp --noroot --caps.drop=all --rlimit-cpu=10 ")
		cmdBuilder.WriteString(binPath)
	} else {
		log.Println("Warning: Running without sandboxing")
		cmdBuilder.WriteString(binPath)
	}

	runCmd = exec.CommandContext(runCtx, "/bin/sh", "-c", cmdBuilder.String())
	runCmd.Dir = tmpDir

	var inputData string
	if len(req.InputLines) > 0 {
		inputData = strings.Join(req.InputLines, "\n") + "\n"
		log.Printf("Using input lines: %v", req.InputLines)
	} else if req.Input != "" {
		inputData = req.Input
		if !strings.HasSuffix(inputData, "\n") {
			inputData += "\n"
		}
		log.Printf("Using single input: %q", req.Input)
	}

	if inputData != "" {
		runCmd.Stdin = strings.NewReader(inputData)
	}

	runOut, err := runCmd.CombinedOutput()

	if err != nil {
		log.Printf("Execution failed: %v\nOutput: %s", err, string(runOut))
		errorMsg := string(runOut)
		if len(errorMsg) == 0 {
			errorMsg = err.Error()
		}

		errorAnalysis, analysisErr := ai.GetErrorAnalysis(req.Code, errorMsg)
		if analysisErr != nil {
			log.Printf("Error analysis failed: %v", analysisErr)
			errorAnalysis = "===Analysis===\n# Erro de Execução\n\nO programa compilou com sucesso, mas encontrou um erro durante a execução. Verifique a divisão por zero, acesso a memória inválida, ou loops infinitos."
		}

		return models.CompileResponse{
			Error:    errorMsg,
			Analysis: errorAnalysis,
		}
	}

	analysis, err := ai.GetAIAnalysis(req.Code, string(runOut))
	if err != nil {
		log.Printf("AI analysis failed: %v", err)
		analysis = "===Analysis===\n# Análise do Código\n\nDesculpe, não foi possível gerar a análise detalhada do código neste momento. O programa compilou e executou com sucesso."
	}

	log.Printf("Program executed successfully. Output length: %d", len(string(runOut)))
	return models.CompileResponse{
		Output:   string(runOut),
		Analysis: analysis,
	}
}
