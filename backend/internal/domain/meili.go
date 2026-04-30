package domain

type MeilisearchResponse struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Text        string `json:"text"`
	Timestamp   string `json:"timestamp"`
}
