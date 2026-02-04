package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func getOllamaURL() string {
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		return "http://localhost:11434"
	}
	return url
}

func getOllamaModel() string {
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		return "llama3.2:1b"
	}
	return model
}

func getAIProvider() string {
	provider := os.Getenv("AI_PROVIDER")
	if provider == "" {
		return "ollama"
	}
	return provider
}

func GetAIAnalysis(code string, output string) (string, error) {
	prompt := fmt.Sprintf(`Você é um professor de programação C. Analise o código abaixo e responda em português.

CÓDIGO:
%s

SAÍDA DO PROGRAMA:
%s

RESPONDA EXATAMENTE NESTE FORMATO (use ## para cada seção):

## Resumo
Uma frase descrevendo o que o programa faz.

## Estrutura
Explique a estrutura e organização do código em 2-3 linhas.

## Funções
Liste as funções/bibliotecas usadas e para que servem.

## Fluxo
Explique passo a passo como o programa executa.

## Análise da Saída
Comente sobre a saída do programa. Se estiver correta ou se indica algum problema lógico.

## Melhorias
Liste sugestões de melhoria se for necessario, não liste se não for necessario.

## Dicas
Uma dica educacional para o estudante.`, code, output)

	return callAI(prompt)
}

func GetErrorAnalysis(code string, errorMessage string) (string, error) {
	prompt := fmt.Sprintf(`Você é um professor de programação C. Analise o erro abaixo e responda em português.

CÓDIGO:
%s

ERRO:
%s

RESPONDA EXATAMENTE NESTE FORMATO (use ## para cada seção):

## Erro
Qual é o erro em uma frase simples.

## Causa
Por que esse erro aconteceu.

## Solução
Como corrigir o erro com exemplo de código.

## Conceito
Explique o conceito de C relacionado ao erro.

## Dicas
Como evitar esse erro no futuro.`, code, errorMessage)

	return callAI(prompt)
}

type GradingResult struct {
	Passed   bool   `json:"passed"`
	Feedback string `json:"feedback"`
}

func GetGradingAnalysis(code string, output string, expectedOutput string) (GradingResult, error) {
	prompt := fmt.Sprintf(`Você é um professor de programação C. Compare a saída do programa do aluno com a saída esperada.
Se a saída for funcionalmente igual (ignorando espaços em branco no final ou diferenças minúsculas de formatação não críticas), considere como correto.
Analise também a qualidade do código.

CODIGO:
%s

SAIDA REAL:
%s

SAIDA ESPERADA:
%s

RESPONDA APENAS UM JSON VÁLIDO (sem markdown, sem explicações extras) neste formato:
{
	"passed": boolean,
	"feedback": "string explicando o resultado e dicas"
}`, code, output, expectedOutput)

	response, err := callAI(prompt)
	if err != nil {
		return GradingResult{}, err
	}

	var result GradingResult

	if len(response) > 3 && response[:3] == "```" {
	}

	cleanResponse := response

	cleanResponse = removeMarkdown(cleanResponse)

	if err := json.Unmarshal([]byte(cleanResponse), &result); err != nil {
		return GradingResult{Passed: false, Feedback: "Erro ao processar resposta da IA: " + response}, nil
	}

	return result, nil
}

func callAI(prompt string) (string, error) {
	provider := getAIProvider()

	switch provider {
	case "groq":
		return callGroqAPI(prompt)
	default:
		return callOllamaAPI(prompt)
	}
}

func callOllamaAPI(prompt string) (string, error) {
	payload := map[string]interface{}{
		"model":       getOllamaModel(),
		"system":      "Você é um professor experiente de programação C, explicando conceitos para um aluno de forma didática e clara.",
		"prompt":      prompt,
		"stream":      false,
		"temperature": 0.3,
		"top_p":       0.9,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	ollamaURL := getOllamaURL() + "/api/generate"
	resp, err := http.Post(ollamaURL,
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

func removeMarkdown(text string) string {
	re := regexp.MustCompile("(?s)```(?:json)?(.*?)```")
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(text)
}

func callGroqAPI(prompt string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}

	url := "https://api.groq.com/openai/v1/chat/completions"

	payload := map[string]interface{}{
		"model": "llama-3.1-8b-instant",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Você é um professor experiente de programação C, explicando conceitos para um aluno de forma didática e clara. Analise todas as partes do codigo e procure por possiveis problemas de logica. Informe o usuario se encontrar. Seja rigido com o aluno, não elogie o codigo se ele estiver errado ou mal feito.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
		"max_tokens":  1024,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling Groq API: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if errObj, ok := result["error"].(map[string]interface{}); ok {
		return "", fmt.Errorf("Groq API error: %v", errObj["message"])
	}

	choices, ok := result["choices"].([]any)
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices in Groq response")
	}

	choice := choices[0].(map[string]interface{})
	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message in Groq response")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("no content in Groq response")
	}

	return content, nil
}
