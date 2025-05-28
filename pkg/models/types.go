package models

// CompileResult represents the result of a compilation
type CompileResult struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// CodeAnalysis represents the AI analysis of the code
type CodeAnalysis struct {
	Elements    []CodeElement `json:"elements"`
	Suggestions []string      `json:"suggestions"`
	AIAnalysis  string        `json:"aiAnalysis"`
}

// CodeElement represents a single element in the code analysis
type CodeElement struct {
	Element     string `json:"element"`
	Description string `json:"description"`
}

// CompileRequest represents the request structure for compilation
type CompileRequest struct {
	Code        string   `json:"code" binding:"required"`
	Input       string   `json:"input,omitempty"`       // Single input string
	InputLines  []string `json:"inputLines,omitempty"`  // Multiple input lines
	Interactive bool     `json:"interactive,omitempty"` // Whether program expects interactive input
	TimeoutSecs int      `json:"timeoutSecs,omitempty"` // Custom timeout (max 30 seconds)
}

// CompileResponse represents the response structure after compilation
type CompileResponse struct {
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Analysis string `json:"analysis,omitempty"`
}
