package models

type CompileRequest struct {
	Code        string   `json:"code"`
	InputLines  []string `json:"input_lines,omitempty"`
	Input       string   `json:"input,omitempty"`
	TimeoutSecs int      `json:"timeout_secs,omitempty"`
}

type CompileResponse struct {
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Analysis string `json:"analysis,omitempty"`
}
