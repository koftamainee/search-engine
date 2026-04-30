package search

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/http-server/middleware/auth"
	"github.com/koftamainee/search-engine/backend/internal/storage"
	"github.com/meilisearch/meilisearch-go"
)

var (
	ErrEmptyQuery    = errors.New("empty query")
	ErrInvalidLimit  = errors.New("invalid limit")
	ErrInvalidOffset = errors.New("invalid offset")
	ErrSearchFailed  = errors.New("search failed")
)

type Service struct {
	index          meilisearch.IndexManager
	historyStorage storage.HistoryStorage
}

func New(index meilisearch.IndexManager, historyStorage storage.HistoryStorage) *Service {
	return &Service{index: index, historyStorage: historyStorage}
}

func (s *Service) Search(ctx context.Context, request domain.SearchRequest) (domain.SearchResponse, error) {
	var result domain.SearchResponse

	if len(request.Query) == 0 {
		return domain.SearchResponse{}, ErrEmptyQuery
	}
	if request.Num <= 0 || request.Num > 50 {
		return domain.SearchResponse{}, ErrInvalidLimit
	}
	if request.Offset < 0 {
		return domain.SearchResponse{}, ErrInvalidOffset
	}
	request.Query = strings.TrimSpace(request.Query)

	if userVal := ctx.Value(auth.UserContextKey); userVal != nil {
		if user, ok := userVal.(*domain.User); ok {
			entry := domain.SearchHistory{
				ID:        uuid.New(),
				UserID:    user.ID,
				Query:     request.Query,
				CreatedAt: time.Now(),
			}

			err := s.historyStorage.Create(ctx, &entry)
			if err != nil {
				log.Printf("failed to save search history: %v", err)
			}
		}
	}

	res, err := s.index.SearchWithContext(ctx, request.Query, &meilisearch.SearchRequest{
		Offset: int64(request.Offset),
		Limit:  int64(request.Num),
	})
	if err != nil {
		return domain.SearchResponse{}, ErrSearchFailed
	}

	hits := make([]domain.SearchResult, 0, len(res.Hits))

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

		var translated domain.SearchResult

		translated.ID = item.ID
		translated.URL = item.URL
		translated.Title = item.Title
		translated.Description = item.Description

		hits = append(hits, translated)
	}

	result = domain.SearchResponse{
		Query: request.Query,
		Hits:  hits,
		Total: int(res.EstimatedTotalHits),

		Num:    request.Num,
		Offset: request.Offset,
	}

	return result, nil
}
