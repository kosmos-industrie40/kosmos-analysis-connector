package connection

// AnalysisAll query all analysis results
type AnalysisAll struct {
	ResultID int    `json:"resultId"`
	Data     int    `json:"data"`
	Machine  string `json:"machine"`
}

// AnalysisMsg query a specific analysis rsult
type AnalysisMsg struct {
	Body struct {
		From      string `json:"from"`
		Timestamp string `json:"timestamp"`
		Model     struct {
			URL string `json:"url"`
			Tag string `json:"tag"`
		} `json:"model"`
		Type       string      `json:"type"`
		Calculated interface{} `json:"calculated"`
	} `json:"body"`
	Signature interface{} `json:"signature"`
}
