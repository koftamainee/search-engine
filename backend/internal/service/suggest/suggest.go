package suggest

import (
	"context"
	"errors"

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

	// TODO: write results to suggestions
	suggestions = append(suggestions, domain.Suggest{Type: "query", Data: "Golang tutorial free without SMS"})
	suggestions = append(suggestions, domain.Suggest{Type: "query", Data: "Cox, Little, O'Shea ch2"})
	suggestions = append(suggestions, domain.Suggest{Type: "query", Data: "RECONSTRUCT WHAT"})
	suggestions = append(suggestions, domain.Suggest{Type: "query", Data: "random hardcoded data"})

	return domain.SuggestResponse{
		Query:       request.Query,
		Suggestions: suggestions,
	}, nil
}
