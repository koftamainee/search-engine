package domain

type SearchRequest struct {
	Query  string `json:"query"`
	Num    int    `json:"num"`
	Offset int    `json:"offset"`
}

type SearchResult struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type SearchResponse struct {
	Query  string         `json:"query"`
	Hits   []SearchResult `json:"hits"`
	Total  int            `json:"total"`
	Num    int            `json:"num"`
	Offset int            `json:"offset"`
}
