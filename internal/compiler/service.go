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

	security.DefaultManager.SetWorkspaceDir(tmpDir, false)
	compileCmd, err := security.DefaultManager.CreateSecureCommand(ctx, "gcc", srcPath, "-o", binPath, "-Wall", "-Wextra")
	if err != nil {
		log.Printf("Failed to create secure compile command: %v", err)
		return models.CompileResponse{Error: "server security configuration error creating compile sandbox"}
	}
	compileCmd.Dir = tmpDir
	log.Printf("Compiling with command: gcc %s -o %s -Wall -Wextra", srcPath, binPath)
	compileOut, err := compileCmd.CombinedOutput()
	security.DefaultManager.CleanupContainer()

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

	if err := security.DefaultManager.ValidateExecutable(binPath); err != nil {
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

	security.DefaultManager.SetWorkspaceDir(tmpDir, true)

	runCmd, err := security.DefaultManager.CreateSecureCommand(runCtx, binPath)
	if err != nil {
		log.Printf("Failed to create secure run command: %v", err)
		return models.CompileResponse{Error: "server security configuration error creating execution sandbox"}
	}
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
	security.DefaultManager.CleanupContainer()

	if err != nil {
		log.Printf("Execution failed: %v\nOutput: %s", err, string(runOut))
		errorMsg := string(runOut)

		if exitErr, ok := err.(*exec.ExitError); ok {
			switch exitErr.ExitCode() {
			case 139:
				errorMsg += "\r\nSegmentation fault (core dumped)"
			case 136:
				errorMsg += "\r\nFloating point exception (core dumped)"
			case 137:
				errorMsg += "\r\nKilled"
			}
		}

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
