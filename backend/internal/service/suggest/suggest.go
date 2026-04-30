package suggest

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/http-server/middleware/auth"
	"github.com/koftamainee/search-engine/backend/internal/storage"
	"github.com/meilisearch/meilisearch-go"
)

var (
	ErrInvalidLimit  = errors.New("invalid limit")
	ErrSuggestFailed = errors.New("suggest failed")
)

type Service struct {
	index   meilisearch.IndexManager
	history storage.HistoryStorage
}

func New(index meilisearch.IndexManager, historyStorage storage.HistoryStorage) *Service {
	return &Service{index: index, history: historyStorage}
}

func (s *Service) Suggest(ctx context.Context, request domain.SuggestRequest) (domain.SuggestResponse, error) {
	limit := request.Limit
	if limit <= 0 || limit > 50 {
		return domain.SuggestResponse{}, ErrInvalidLimit
	}

	if request.Query == "" {
		userVal := ctx.Value(auth.UserContextKey)
		if userVal != nil {
			user, ok := userVal.(*domain.User)
			if ok {
				history, err := s.history.GetByUserID(ctx, user.ID, limit, 0)
				if err != nil {
					return domain.SuggestResponse{}, ErrSuggestFailed
				}

				seen := make(map[string]struct{})
				suggestions := make([]domain.Suggest, 0, len(*history))
				for _, h := range *history {
					if _, exists := seen[h.Query]; exists {
						continue
					}
					seen[h.Query] = struct{}{}
					suggestions = append(suggestions, domain.Suggest{
						Type: "history",
						Data: h.Query,
					})
				}
				return domain.SuggestResponse{
					Query:       request.Query,
					Suggestions: suggestions,
				}, nil
			}
		}
	}

	searchReq := &meilisearch.SearchRequest{
		Limit: int64(limit),
	}
	res, err := s.index.SearchWithContext(ctx, request.Query, searchReq)
	if err != nil {
		return domain.SuggestResponse{}, ErrSuggestFailed
	}

	suggestions := make([]domain.Suggest, 0, len(res.Hits))

	for _, h := range res.Hits {
		b, err := json.Marshal(h)
		if err != nil {
			log.Printf("failed to marshal search result: %v", err)
			continue
		}

		var item domain.MeilisearchResponse
		if err := json.Unmarshal(b, &item); err != nil {
			log.Printf("failed to unmarshal search result: %v", err)
			continue
		}

		var translated domain.Suggest

		translated.Type = "query"
		translated.Data = item.Title

		suggestions = append(suggestions, translated)
	}

	return domain.SuggestResponse{
		Query:       request.Query,
		Suggestions: suggestions,
	}, nil
}
