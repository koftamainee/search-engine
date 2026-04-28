package domain

type SuggestRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type Suggest struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type SuggestResponse struct {
	Query       string    `json:"query"`
	Suggestions []Suggest `json:"suggestions"`
}
