package search

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
	"github.com/koftamainee/search-engine/backend/internal/service/search"
)

func New(searchService *search.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		query := r.URL.Query()

		req := domain.SearchRequest{
			Query:  query.Get("q"),
			Num:    10,
			Offset: 0,
		}

		if numStr := query.Get("num"); numStr != "" {
			num, err := strconv.Atoi(numStr)
			if err == nil {
				req.Num = num
			}
		}

		if offsetStr := query.Get("offset"); offsetStr != "" {
			offset, err := strconv.Atoi(offsetStr)
			if err == nil {
				req.Offset = offset
			}
		}

		res, err := searchService.Search(ctx, req)
		if err != nil {
			switch {
			case errors.Is(err, search.ErrEmptyQuery):
				response.BadRequest(w, "empty query")
			case errors.Is(err, search.ErrInvalidLimit):
				response.BadRequest(w, "invalid limit")
			case errors.Is(err, search.ErrInvalidOffset):
				response.BadRequest(w, "invalid offset")
			case errors.Is(err, search.ErrSearchFailed):
				response.Internal(w)

			default:
				response.Internal(w)
			}
			return
		}

		response.OK(w, res)
	}
}
