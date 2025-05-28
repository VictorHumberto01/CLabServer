package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetAIAnalysis generates AI analysis for successful compilation
func GetAIAnalysis(code string) (string, error) {
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

// GetErrorAnalysis generates AI analysis for compilation errors
func GetErrorAnalysis(code string, errorMessage string) (string, error) {
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

// callOllamaAPI makes a request to the Ollama API
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
