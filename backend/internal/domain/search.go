package domain

type SearchRequest struct {
	Query  string `json:"query"`
	Num    int    `json:"num"`
	Offset int    `json:"offset"`
}

type SearchResult struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
	Data  any     `json:"data"`
}

type SearchResponse struct {
	Query     string         `json:"query"`
	Hits      []SearchResult `json:"hits"`
	Num       int            `json:"num"`
	Total     int            `json:"total"`
	Start     int            `json:"start"`
	NextStart int            `json:"next_start,omitempty"`
}
