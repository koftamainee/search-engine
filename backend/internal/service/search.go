package service

import (
	"context"
	"errors"
	"strings"

	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/meilisearch/meilisearch-go"
)

var (
	ErrEmptyQuery    = errors.New("empty query")
	ErrInvalidLimit  = errors.New("invalid limit")
	ErrInvalidOffset = errors.New("invalid offset")
	ErrSearchFailed  = errors.New("search failed")
)

type SearchService struct {
	index meilisearch.IndexManager
}

func NewSearchService(index meilisearch.IndexManager) *SearchService {
	return &SearchService{index: index}
}

func (s *SearchService) Search(ctx context.Context, request domain.SearchRequest) (domain.SearchResponse, error) {
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

	res, err := s.index.SearchWithContext(ctx, request.Query, &meilisearch.SearchRequest{
		Limit:  int64(request.Num),
		Offset: int64(request.Offset),
	})
	if err != nil {
		return domain.SearchResponse{}, ErrSearchFailed
	}

	hits := make([]domain.SearchResult, 0, len(res.Hits))

	// TODO: write results to hits

	result = domain.SearchResponse{
		Query: request.Query,
		Hits:  hits,
		Total: int(res.EstimatedTotalHits),

		Num:   request.Num,
		Start: request.Offset,
	}

	return result, nil
}
